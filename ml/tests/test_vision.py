"""Unit tests for BioCLIP / SpeciesNet wildlife roadkill photo classifier and ONNX export."""

from __future__ import annotations

import io
import os
from pathlib import Path
import numpy as np
import pytest
from PIL import Image
import onnxruntime as ort

from corridor_ml.vision.classifier import (
    SpeciesClassifier,
    TARGET_SPECIES,
    preprocess_image,
)
from corridor_ml.vision.export_onnx import export_species_model


def _create_sample_image_bytes(format: str = "JPEG", size: tuple[int, int] = (300, 200), color=(180, 120, 60)) -> bytes:
    """Helper to generate valid synthetic image bytes."""
    img = Image.new("RGB", size, color=color)
    buf = io.BytesIO()
    img.save(buf, format=format)
    return buf.getvalue()


def _create_rgba_png_bytes(size: tuple[int, int] = (250, 250)) -> bytes:
    """Helper to generate RGBA PNG bytes with alpha channel."""
    img = Image.new("RGBA", size, color=(100, 150, 200, 128))
    buf = io.BytesIO()
    img.save(buf, format="PNG")
    return buf.getvalue()


class TestImagePreprocessing:
    """Tests for image loading, resizing, RGB conversion, and normalization."""

    def test_preprocess_jpeg_shape_and_dtype(self):
        jpeg_bytes = _create_sample_image_bytes(format="JPEG", size=(640, 480))
        tensor = preprocess_image(jpeg_bytes)

        assert isinstance(tensor, np.ndarray)
        assert tensor.shape == (1, 3, 224, 224)
        assert tensor.dtype == np.float32

    def test_preprocess_png_rgba_conversion(self):
        png_bytes = _create_rgba_png_bytes(size=(128, 128))
        tensor = preprocess_image(png_bytes)

        assert tensor.shape == (1, 3, 224, 224)
        assert tensor.dtype == np.float32

    def test_preprocess_normalization_stats(self):
        # Pure white image: (255, 255, 255) -> 1.0 in float
        # Normalized: (1.0 - mean) / std
        white_img = Image.new("RGB", (224, 224), color=(255, 255, 255))
        buf = io.BytesIO()
        white_img.save(buf, format="JPEG")
        tensor = preprocess_image(buf.getvalue())

        expected_ch0 = (1.0 - 0.485) / 0.229
        expected_ch1 = (1.0 - 0.456) / 0.224
        expected_ch2 = (1.0 - 0.406) / 0.225

        # Check approximately equal (JPEG compression may slightly shift 255)
        assert np.isclose(tensor[0, 0, :, :].mean(), expected_ch0, atol=0.05)
        assert np.isclose(tensor[0, 1, :, :].mean(), expected_ch1, atol=0.05)
        assert np.isclose(tensor[0, 2, :, :].mean(), expected_ch2, atol=0.05)

    @pytest.mark.parametrize(
        "invalid_input",
        [
            b"",
            b"corrupt non-image header bytes",
            b"\x89PNG\r\n\x1a\ncorrupt_content",
            b"\xff\xd8\xff\xe0truncated_jpeg",
        ],
    )
    def test_preprocess_invalid_bytes_raises_value_error(self, invalid_input):
        with pytest.raises(ValueError, match="(?i)invalid|corrupt"):
            preprocess_image(invalid_input)


class TestSpeciesClassifier:
    """Tests for SpeciesClassifier inference and predictions."""

    def test_target_species_classes(self):
        assert len(TARGET_SPECIES) == 8
        expected = [
            "mule_deer",
            "white_tailed_deer",
            "elk",
            "moose",
            "pronghorn",
            "black_bear",
            "raccoon",
            "coyote",
        ]
        assert TARGET_SPECIES == expected

    def test_predict_top_k_default(self):
        classifier = SpeciesClassifier()
        img_bytes = _create_sample_image_bytes(format="JPEG")

        predictions = classifier.predict_top_k(img_bytes, k=3)

        assert isinstance(predictions, list)
        assert len(predictions) == 3

        prev_prob = 1.0
        for species, prob in predictions:
            assert isinstance(species, str)
            assert species in TARGET_SPECIES
            assert isinstance(prob, float)
            assert 0.0 <= prob <= 1.0
            assert prob <= prev_prob  # Sorted descending
            prev_prob = prob

    def test_predict_top_k_custom_k(self):
        classifier = SpeciesClassifier()
        img_bytes = _create_sample_image_bytes(format="PNG")

        top_1 = classifier.predict_top_k(img_bytes, k=1)
        assert len(top_1) == 1

        top_8 = classifier.predict_top_k(img_bytes, k=8)
        assert len(top_8) == 8
        total_prob = sum(p for _, p in top_8)
        assert np.isclose(total_prob, 1.0, atol=1e-4)

    def test_predict_top_k_invalid_image_raises_value_error(self):
        classifier = SpeciesClassifier()
        with pytest.raises(ValueError, match="(?i)invalid|corrupt"):
            classifier.predict_top_k(b"not an image", k=3)


class TestONNXExport:
    """Tests for ONNX model export, dynamic int8 quantization, and size limits."""

    def test_export_onnx_quantized_under_35mb(self, tmp_path: Path):
        classifier = SpeciesClassifier()
        out_onnx = tmp_path / "species_classifier_int8.onnx"

        classifier.export_onnx(str(out_onnx), quantize_int8=True)

        assert out_onnx.exists()
        size_bytes = os.path.getsize(out_onnx)
        size_mb = size_bytes / (1024 * 1024)

        # Must be strictly under 35 MB
        assert size_mb < 35.0
        assert size_bytes > 1000  # Non-empty, valid model

    def test_export_onnx_fp32_identical_logits(self, tmp_path: Path):
        classifier = SpeciesClassifier()
        fp32_path = tmp_path / "species_classifier_fp32.onnx"

        classifier.export_onnx(str(fp32_path), quantize_int8=False)

        assert fp32_path.exists()
        session = ort.InferenceSession(str(fp32_path))

        test_tensor = np.random.randn(1, 3, 224, 224).astype(np.float32)
        onnx_logits = session.run(None, {"input": test_tensor})[0]
        classifier_logits = classifier.predict_logits(test_tensor)

        np.testing.assert_allclose(onnx_logits, classifier_logits, rtol=1e-4, atol=1e-4)

    def test_export_onnx_int8_near_identical_logits(self, tmp_path: Path):
        classifier = SpeciesClassifier()
        fp32_path = tmp_path / "model_fp32.onnx"
        int8_path = tmp_path / "model_int8.onnx"

        classifier.export_onnx(str(fp32_path), quantize_int8=False)
        classifier.export_onnx(str(int8_path), quantize_int8=True)

        sess_fp32 = ort.InferenceSession(str(fp32_path))
        sess_int8 = ort.InferenceSession(str(int8_path))

        test_tensor = np.random.randn(1, 3, 224, 224).astype(np.float32)
        logits_fp32 = sess_fp32.run(None, {"input": test_tensor})[0]
        logits_int8 = sess_int8.run(None, {"input": test_tensor})[0]

        # Logits should be close within int8 quantization error margin
        np.testing.assert_allclose(logits_fp32, logits_int8, atol=0.5)

        # Top predicted class index should match
        assert np.argmax(logits_fp32) == np.argmax(logits_int8)

    def test_export_species_model_utility(self, tmp_path: Path):
        out_path = tmp_path / "utility_export.onnx"
        res_path = export_species_model(out_path, quantize_int8=True)

        assert Path(res_path).exists()
        size_mb = os.path.getsize(res_path) / (1024 * 1024)
        assert size_mb < 35.0
