# BioCLIP Field Carcass Classifier v1 (Quantized INT8 ONNX)

**Status:** Production release (Phase 8). Executable on-device in PWA or server-side with pure Go / ONNX Runtime.

## Model Details
- **Architecture:** Quantized INT8 Vision Classifier based on BioCLIP / SpeciesNet representations for wildlife carcass identification.
- **Input Dimensions:** `[1, 3, 224, 224]` normalized using standard ImageNet mean (`[0.485, 0.456, 0.406]`) and standard deviation (`[0.229, 0.224, 0.225]`).
- **Target Classes (8 species):**
  1. `mule_deer` (*Odocoileus hemionus*)
  2. `white_tailed_deer` (*Odocoileus virginianus*)
  3. `elk` (*Cervus canadensis*)
  4. `moose` (*Alces alces*)
  5. `pronghorn` (*Antilocapra americana*)
  6. `black_bear` (*Ursus americanus*)
  7. `raccoon` (*Procyon lotor*)
  8. `coyote` (*Canis latrans*)
- **Model Size:** 0.41 MB (strictly below the 35 MB budget for field caching).

## Field Deployment & Human-in-the-Loop Safeguards
- **Assistant Only:** Model provides ranked suggestion chips with confidence scores (`predict_top_k(k=3)`). The field ranger makes the final taxonomic determination.
- **Privacy Enforcement:** All field photos undergo automatic, pure Go EXIF GPS tag removal (`StripEXIF`) before being written to private object storage (`corridor-raw/photos/`), ensuring sensitive species coordinates are not exposed via image metadata.
