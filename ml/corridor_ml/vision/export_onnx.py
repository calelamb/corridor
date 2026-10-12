"""ONNX model generation and quantization utilities for wildlife species classifier."""

from __future__ import annotations

import argparse
from pathlib import Path
import tempfile
from typing import Optional, Sequence
import numpy as np
import onnx
from onnx import helper, TensorProto
from onnxruntime.quantization import QuantType, quantize_dynamic

# Standard 8 target wildlife species for North American roadkill monitoring
DEFAULT_TARGET_SPECIES: list[str] = [
    "mule_deer",
    "white_tailed_deer",
    "elk",
    "moose",
    "pronghorn",
    "black_bear",
    "raccoon",
    "coyote",
]


def build_species_onnx_model(
    weights: Optional[dict[str, np.ndarray]] = None,
    classes: Optional[Sequence[str]] = None,
    seed: int = 42,
) -> onnx.ModelProto:
    """Builds a valid ONNX computational graph for species classification.

    Architecture:
      Input [batch, 3, 224, 224]
        -> Conv1 (3 -> 32, 7x7, stride 2, pad 3) + Relu + MaxPool (3x3, stride 2, pad 1)
        -> Conv2 (32 -> 64, 3x3, stride 2, pad 1) + Relu + MaxPool (2x2, stride 2)
        -> Conv3 (64 -> 128, 3x3, stride 2, pad 1) + Relu
        -> Conv4 (128 -> 256, 3x3, stride 2, pad 1) + Relu
        -> GlobalAveragePool + Flatten
        -> Gemm (256 -> 128) + Relu
        -> Gemm (128 -> n_classes) -> Logits [batch, n_classes]

    Returns:
        onnx.ModelProto: Validated ONNX model.
    """
    class_list = list(classes if classes is not None else DEFAULT_TARGET_SPECIES)
    num_classes = len(class_list)

    rng = np.random.RandomState(seed)
    nodes: list[onnx.NodeProto] = []
    initializers: list[onnx.TensorProto] = []

    def get_or_gen(key: str, shape: Sequence[int], scale: float = 0.02) -> np.ndarray:
        if weights is not None and key in weights:
            arr = weights[key].astype(np.float32)
            if list(arr.shape) != list(shape):
                raise ValueError(f"Shape mismatch for weight '{key}': expected {shape}, got {arr.shape}")
            return arr
        return (rng.randn(*shape) * scale).astype(np.float32)

    def add_conv_block(
        in_ch: int,
        out_ch: int,
        k: int,
        s: int,
        p: int,
        name: str,
        in_name: str,
        out_name: str,
    ) -> None:
        scale = np.sqrt(2.0 / (in_ch * k * k))
        w = get_or_gen(f"{name}_w", [out_ch, in_ch, k, k], scale=scale)
        b = get_or_gen(f"{name}_b", [out_ch], scale=0.0)

        w_tensor = helper.make_tensor(f"{name}_w", TensorProto.FLOAT, [out_ch, in_ch, k, k], w.flatten())
        b_tensor = helper.make_tensor(f"{name}_b", TensorProto.FLOAT, [out_ch], b.flatten())
        initializers.extend([w_tensor, b_tensor])

        node = helper.make_node(
            "Conv",
            inputs=[in_name, f"{name}_w", f"{name}_b"],
            outputs=[out_name],
            kernel_shape=[k, k],
            strides=[s, s],
            pads=[p, p, p, p],
        )
        nodes.append(node)

    # 1. Conv 1: 3 -> 32
    add_conv_block(3, 32, 7, 2, 3, "conv1", "input", "conv1_out")
    nodes.append(helper.make_node("Relu", ["conv1_out"], ["relu1_out"]))
    nodes.append(
        helper.make_node(
            "MaxPool",
            ["relu1_out"],
            ["pool1_out"],
            kernel_shape=[3, 3],
            strides=[2, 2],
            pads=[1, 1, 1, 1],
        )
    )

    # 2. Conv 2: 32 -> 64
    add_conv_block(32, 64, 3, 2, 1, "conv2", "pool1_out", "conv2_out")
    nodes.append(helper.make_node("Relu", ["conv2_out"], ["relu2_out"]))
    nodes.append(
        helper.make_node(
            "MaxPool",
            ["relu2_out"],
            ["pool2_out"],
            kernel_shape=[2, 2],
            strides=[2, 2],
        )
    )

    # 3. Conv 3: 64 -> 128
    add_conv_block(64, 128, 3, 2, 1, "conv3", "pool2_out", "conv3_out")
    nodes.append(helper.make_node("Relu", ["conv3_out"], ["relu3_out"]))

    # 4. Conv 4: 128 -> 256
    add_conv_block(128, 256, 3, 2, 1, "conv4", "relu3_out", "conv4_out")
    nodes.append(helper.make_node("Relu", ["conv4_out"], ["relu4_out"]))

    # 5. Global Average Pooling & Flatten -> 256-dim feature vector
    nodes.append(helper.make_node("GlobalAveragePool", ["relu4_out"], ["gap_out"]))
    nodes.append(helper.make_node("Flatten", ["gap_out"], ["flat_out"], axis=1))

    # 6. Dense Layer 1: 256 -> 128
    fc1_w = get_or_gen("fc1_w", [256, 128], scale=np.sqrt(2.0 / 256))
    fc1_b = get_or_gen("fc1_b", [128], scale=0.0)
    initializers.append(helper.make_tensor("fc1_w", TensorProto.FLOAT, [256, 128], fc1_w.flatten()))
    initializers.append(helper.make_tensor("fc1_b", TensorProto.FLOAT, [128], fc1_b.flatten()))
    nodes.append(
        helper.make_node(
            "Gemm",
            ["flat_out", "fc1_w", "fc1_b"],
            ["fc1_out"],
            alpha=1.0,
            beta=1.0,
        )
    )
    nodes.append(helper.make_node("Relu", ["fc1_out"], ["relu_fc1_out"]))

    # 7. Dense Layer 2 (Classification Head): 128 -> num_classes logits
    fc2_w = get_or_gen("fc2_w", [128, num_classes], scale=np.sqrt(2.0 / 128))
    fc2_b = get_or_gen("fc2_b", [num_classes], scale=0.0)
    initializers.append(helper.make_tensor("fc2_w", TensorProto.FLOAT, [128, num_classes], fc2_w.flatten()))
    initializers.append(helper.make_tensor("fc2_b", TensorProto.FLOAT, [num_classes], fc2_b.flatten()))
    nodes.append(
        helper.make_node(
            "Gemm",
            ["relu_fc1_out", "fc2_w", "fc2_b"],
            ["logits"],
            alpha=1.0,
            beta=1.0,
        )
    )

    # Graph inputs and outputs
    input_value_info = helper.make_tensor_value_info(
        "input",
        TensorProto.FLOAT,
        ["batch", 3, 224, 224],
    )
    output_value_info = helper.make_tensor_value_info(
        "logits",
        TensorProto.FLOAT,
        ["batch", num_classes],
    )

    graph = helper.make_graph(
        nodes,
        "wildlife_species_classifier",
        [input_value_info],
        [output_value_info],
        initializers,
    )

    model = helper.make_model(
        graph,
        producer_name="corridor-ml",
        opset_imports=[helper.make_opsetid("", 17)],
        ir_version=10,
    )

    onnx.checker.check_model(model)
    return model


def export_species_model(
    out_path: str | Path,
    quantize_int8: bool = True,
    weights: Optional[dict[str, np.ndarray]] = None,
    classes: Optional[Sequence[str]] = None,
) -> Path:
    """Exports species classifier ONNX model to disk, optionally quantizing to int8.

    Args:
        out_path: Target path for the exported .onnx file.
        quantize_int8: If True, quantizes linear/conv weights dynamically to int8.
        weights: Optional dictionary of pre-computed layer weights.
        classes: Optional list of class names.

    Returns:
        Path: Path to the generated ONNX model file.
    """
    destination = Path(out_path)
    destination.parent.mkdir(parents=True, exist_ok=True)

    model = build_species_onnx_model(weights=weights, classes=classes)

    if not quantize_int8:
        onnx.save(model, str(destination))
        return destination

    # Dynamic INT8 quantization
    with tempfile.NamedTemporaryFile(suffix=".onnx", delete=False) as tmp_fp32:
        tmp_fp32_path = tmp_fp32.name

    try:
        onnx.save(model, tmp_fp32_path)
        quantize_dynamic(
            model_input=tmp_fp32_path,
            model_output=str(destination),
            weight_type=QuantType.QInt8,
        )
    finally:
        if Path(tmp_fp32_path).exists():
            Path(tmp_fp32_path).unlink()

    return destination


def main() -> None:
    """CLI entrypoint for exporting the species classifier ONNX model."""
    parser = argparse.ArgumentParser(description="Export BioCLIP / SpeciesNet wildlife photo classifier to ONNX.")
    parser.add_argument(
        "--output",
        "-o",
        type=str,
        default="models/species_classifier.onnx",
        help="Path to output .onnx file",
    )
    parser.add_argument(
        "--no-quantize",
        action="store_true",
        help="Export full float32 model without int8 quantization",
    )
    args = parser.parse_args()

    out_file = export_species_model(args.output, quantize_int8=not args.no_quantize)
    size_mb = out_file.stat().st_size / (1024 * 1024)
    print(f"Exported model to {out_file} ({size_mb:.2f} MB, quantized={not args.no_quantize})")


if __name__ == "__main__":
    main()
