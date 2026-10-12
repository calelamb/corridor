"""BioCLIP / SpeciesNet wildlife roadkill photo classifier and inference engine."""

from __future__ import annotations

import io
from pathlib import Path
import shutil
import tempfile
from typing import Optional, Sequence
import numpy as np
import onnx
from onnxruntime.quantization import QuantType, quantize_dynamic
import onnxruntime as ort
from PIL import Image

from corridor_ml.vision.export_onnx import (
    DEFAULT_TARGET_SPECIES,
    build_species_onnx_model,
    export_species_model,
)

# Standard North American wildlife carcass monitoring targets
TARGET_SPECIES: list[str] = list(DEFAULT_TARGET_SPECIES)

# Standard ImageNet normalization parameters (compatible with BioCLIP / CLIP backbones)
IMAGENET_MEAN: np.ndarray = np.array([0.485, 0.456, 0.406], dtype=np.float32)
IMAGENET_STD: np.ndarray = np.array([0.229, 0.224, 0.225], dtype=np.float32)


def preprocess_image(image_bytes: bytes) -> np.ndarray:
    """Decodes raw JPEG or PNG image bytes and formats them for the vision model.

    Operations:
      1. Validates and decodes image bytes using PIL.
      2. Converts color space to 3-channel RGB.
      3. Resizes image to standard model input resolution (224, 224).
      4. Scales pixel values to [0.0, 1.0] float32.
      5. Normalizes using standard ImageNet mean and std.
      6. Formats as NCHW tensor with leading batch dimension: [1, 3, 224, 224].

    Args:
        image_bytes: Raw bytes representing a JPEG or PNG image.

    Returns:
        np.ndarray: Preprocessed float32 array of shape (1, 3, 224, 224).

    Raises:
        ValueError: If image_bytes is empty, corrupted, or cannot be decoded as an image.
    """
    if not isinstance(image_bytes, (bytes, bytearray)) or len(image_bytes) == 0:
        raise ValueError("Invalid or corrupt image bytes: input is empty or not bytes.")

    try:
        with Image.open(io.BytesIO(image_bytes)) as raw_img:
            # Force decoding of image data to detect corrupt / truncated streams early
            raw_img.load()
            rgb_img = raw_img.convert("RGB")
            resized_img = rgb_img.resize((224, 224), Image.Resampling.BILINEAR)
    except Exception as exc:
        raise ValueError(f"Invalid or corrupt image bytes: failed to decode image: {exc}") from exc

    # Convert to float32 scaled to [0.0, 1.0]
    img_data = np.asarray(resized_img, dtype=np.float32) / 255.0

    # Normalize channels
    normalized = (img_data - IMAGENET_MEAN) / IMAGENET_STD

    # Transpose HWC -> CHW and add batch dimension -> NCHW
    chw = np.transpose(normalized, (2, 0, 1))
    tensor = np.expand_dims(chw, axis=0)

    return tensor.astype(np.float32)


class SpeciesClassifier:
    """Roadkill wildlife carcass classifier for field reporting and verification.

    Targets 8 primary ungulate and predator species:
      - mule_deer
      - white_tailed_deer
      - elk
      - moose
      - pronghorn
      - black_bear
      - raccoon
      - coyote
    """

    TARGET_SPECIES = TARGET_SPECIES

    def __init__(
        self,
        model_path: Optional[str | Path] = None,
        classes: Optional[Sequence[str]] = None,
    ) -> None:
        """Initializes the SpeciesClassifier.

        If model_path is provided, loads the existing ONNX model from disk.
        Otherwise, builds and initializes an in-memory optimized ONNX computational graph.
        """
        self.classes: list[str] = list(classes if classes is not None else TARGET_SPECIES)
        self._model_path: Optional[str] = None
        self._model_proto: Optional[onnx.ModelProto] = None

        if model_path is not None:
            path_str = str(model_path)
            self._session = ort.InferenceSession(path_str)
            self._model_path = path_str
        else:
            self._model_proto = build_species_onnx_model(classes=self.classes)
            # Create session directly from serialized in-memory proto
            self._session = ort.InferenceSession(self._model_proto.SerializeToString())

    def preprocess(self, image_bytes: bytes) -> np.ndarray:
        """Helper to preprocess raw image bytes into model input tensor."""
        return preprocess_image(image_bytes)

    def predict_logits(self, input_tensor: np.ndarray) -> np.ndarray:
        """Runs the ONNX computational graph forward pass and returns raw logits.

        Args:
            input_tensor: Tensor of shape [batch, 3, 224, 224] (float32).

        Returns:
            np.ndarray: Logits array of shape [batch, n_classes].
        """
        if (
            not isinstance(input_tensor, np.ndarray)
            or input_tensor.ndim != 4
            or input_tensor.shape[1:] != (3, 224, 224)
        ):
            raise ValueError(f"Expected 4D input tensor [batch, 3, 224, 224], got {getattr(input_tensor, 'shape', None)}")

        outputs = self._session.run(None, {"input": input_tensor.astype(np.float32)})
        return outputs[0]

    def predict_proba(self, input_tensor: np.ndarray) -> np.ndarray:
        """Calculates normalized class probabilities using numerically stable softmax.

        Args:
            input_tensor: Tensor of shape [batch, 3, 224, 224].

        Returns:
            np.ndarray: Probabilities array of shape [batch, n_classes].
        """
        logits = self.predict_logits(input_tensor)
        if logits.shape[-1] != len(self.classes):
            raise ValueError(f"Model output dimension {logits.shape[-1]} does not match number of classes {len(self.classes)}")

        # Shift logits by max for numerical stability
        shifted = logits - np.max(logits, axis=-1, keepdims=True)
        exp_vals = np.exp(shifted)
        probs = exp_vals / np.sum(exp_vals, axis=-1, keepdims=True)
        return probs

    def predict_top_k(self, image_bytes: bytes, k: int = 3) -> list[tuple[str, float]]:
        """Predicts top-k most likely wildlife species from raw photo bytes.

        Args:
            image_bytes: Raw bytes of JPEG or PNG field photo.
            k: Number of candidate species to return (default: 3).

        Returns:
            list[tuple[str, float]]: List of (species_name, probability) tuples,
                                     sorted descending by probability.
        """
        tensor = self.preprocess(image_bytes)
        probs = self.predict_proba(tensor)[0]

        pairs: list[tuple[str, float]] = [
            (self.classes[i], float(probs[i])) for i in range(len(self.classes))
        ]
        pairs.sort(key=lambda item: item[1], reverse=True)

        k_clamped = max(0, min(k, len(pairs)))
        return pairs[:k_clamped]

    def export_onnx(self, out_path: str | Path, quantize_int8: bool = True) -> Path:
        """Exports the classifier model to an optimized ONNX file on disk.

        Args:
            out_path: Destination path for the exported .onnx file.
            quantize_int8: Whether to apply dynamic 8-bit integer quantization.

        Returns:
            Path: Path to the generated ONNX file.
        """
        destination = Path(out_path)
        destination.parent.mkdir(parents=True, exist_ok=True)

        if self._model_path is not None:
            # Loaded from existing file
            if quantize_int8:
                quantize_dynamic(
                    model_input=self._model_path,
                    model_output=str(destination),
                    weight_type=QuantType.QInt8,
                )
            else:
                shutil.copyfile(self._model_path, destination)
            return destination

        # In-memory model proto
        return export_species_model(
            out_path=destination,
            quantize_int8=quantize_int8,
            model=self._model_proto,
            classes=self.classes,
        )
