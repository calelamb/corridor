"""TreeSHAP attribution engine for per-segment risk driver extraction.

Computes exact tree-based Shapley values (TreeSHAP) on fitted LightGBM models
to explain why each road segment is hazardous. Produces interpretable,
human-actionable risk drivers with semantic directions (e.g. "closer to stream",
"high traffic volume", "winter range overlap").
"""

from __future__ import annotations
from typing import Optional, Sequence, Union
import numpy as np
import pandas as pd
import shap
import lightgbm as lgb


NON_FEATURE_COLUMNS = {
    "segment_id",
    "spatial_block",
    "geometry",
    "target",
    "y",
    "collision_count",
}


def infer_feature_direction(
    feature_name: str,
    value: float,
    reference_value: float,
) -> str:
    """Infers human-interpretable semantic direction of a feature's risk impact.

    Parameters
    ----------
    feature_name : name of the covariate
    value : observed value of the feature for this segment
    reference_value : population baseline/median value of this feature

    Returns
    -------
    str : semantic direction string ("closer", "further", "higher", "lower",
          "present", "absent", "denser", "sparser", "steeper", "gentler")
    """
    fname = feature_name.lower()

    # Distance to geographical entities (e.g. stream, corridor)
    if any(k in fname for k in ("dist_", "distance")):
        return "closer" if value < reference_value else "further"

    # Discrete range overlap / wildlife corridor presence / flags
    if any(k in fname for k in ("overlap", "corridor", "presence", "is_", "has_")):
        return "present" if value >= 0.5 else "absent"

    # Canopy / vegetation density
    if "canopy" in fname:
        return "denser" if value >= reference_value else "sparser"

    # Slope / terrain ruggedness
    if any(k in fname for k in ("slope", "ruggedness", "tri")):
        return "steeper" if value >= reference_value else "gentler"

    # Traffic volume, speed limit, general rates
    return "higher" if value >= reference_value else "lower"


def export_shap_reasons(
    model: lgb.Booster,
    X: pd.DataFrame,
    top_n: int = 3,
    id_col: str = "segment_id",
    only_positive: bool = True,
    reference_values: Optional[dict[str, float]] = None,
) -> pd.DataFrame:
    """Extracts top N risk drivers per segment using TreeSHAP explanations.

    Parameters
    ----------
    model : trained LightGBM Booster
    X : DataFrame containing segment features (and optional segment_id)
    top_n : maximum number of top drivers per segment (default 3)
    id_col : column name containing segment identifiers (default "segment_id")
    only_positive : whether to restrict reasons strictly to positive risk drivers
                    (features increasing expected collision rate)
    reference_values : optional mapping of feature names to baseline reference values;
                       computed as median across X if None.

    Returns
    -------
    pd.DataFrame with columns:
        - segment_id: segment identifier
        - rank: priority rank (1 = strongest positive driver)
        - feature: name of covariate
        - direction: semantic direction ("closer", "higher", "present", etc.)
        - contribution: TreeSHAP attribution value (log-rate shift)
        - feature_value: raw observed feature value
        - impact: formatted human-readable summary
    """
    if not isinstance(model, lgb.Booster):
        raise TypeError(f"Expected lgb.Booster, got {type(model)}")

    feature_names = model.feature_name()
    if not feature_names:
        raise ValueError("Model has no feature names registered.")

    missing_features = [f for f in feature_names if f not in X.columns]
    if missing_features:
        raise ValueError(f"Features missing from input DataFrame: {missing_features}")

    # Prepare feature matrix
    X_features = X[feature_names]
    X_mat = X_features.to_numpy(dtype=np.float64)
    n_samples, n_features = X_mat.shape

    # Segment IDs
    if id_col in X.columns:
        segment_ids = [str(sid) for sid in X[id_col].values]
    else:
        segment_ids = [str(idx) for idx in X.index]

    # Calculate baseline references (median) if not supplied
    if reference_values is None:
        reference_values = {
            f: float(X_features[f].median()) if len(X_features) > 0 else 0.0
            for f in feature_names
        }

    # Run TreeSHAP
    explainer = shap.TreeExplainer(model)
    shap_values = explainer.shap_values(X_mat)
    if isinstance(shap_values, list):
        shap_values = shap_values[0]

    records = []
    for i in range(n_samples):
        seg_id = segment_ids[i]
        row_shap = shap_values[i]
        row_feats = X_mat[i]

        # Filter candidate features
        if only_positive:
            candidate_indices = [
                j for j in range(n_features) if float(row_shap[j]) > 1e-7
            ]
        else:
            candidate_indices = list(range(n_features))

        # Sort candidate features by SHAP contribution descending
        sorted_indices = sorted(
            candidate_indices, key=lambda j: float(row_shap[j]), reverse=True
        )[:top_n]

        for rank, j in enumerate(sorted_indices, start=1):
            feat_name = feature_names[j]
            feat_val = float(row_feats[j])
            shap_val = float(row_shap[j])
            ref_val = reference_values.get(feat_name, 0.0)

            direction = infer_feature_direction(feat_name, feat_val, ref_val)
            sign_str = f"+{shap_val:.2f}" if shap_val >= 0 else f"{shap_val:.2f}"
            impact_str = f"{feat_name} ({direction}, {sign_str})"

            records.append(
                {
                    "segment_id": seg_id,
                    "rank": rank,
                    "feature": feat_name,
                    "direction": direction,
                    "contribution": round(shap_val, 6),
                    "feature_value": round(feat_val, 4),
                    "impact": impact_str,
                }
            )

    columns = [
        "segment_id",
        "rank",
        "feature",
        "direction",
        "contribution",
        "feature_value",
        "impact",
    ]
    if not records:
        return pd.DataFrame(columns=columns)

    return pd.DataFrame(records)[columns]
