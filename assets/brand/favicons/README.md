Favicons and social preview

This folder contains generated favicons and social preview images for the repository.

How to generate

1. Ensure you have one of the following installed:
   - `rsvg-convert` (recommended, from librsvg)
   - ImageMagick `convert`
   - `icotool` (optional, to create multi-size `favicon.ico`)

2. Run the generator from the repo root:

```bash
bash scripts/generate-favicons.sh
```

3. Commit the generated files in `assets/brand/favicons/`.

Placement suggestions

- `assets/brand/favicons/favicon.ico` — place at repo root or reference it in HTML pages.
- `assets/brand/favicons/social-preview.png` — upload to GitHub Repository Settings > Social preview.
- `apple-touch-icon-180.png` — for iOS home-screen icons.

Notes

- The script uses `assets/small-sticky.svg` as the source. Replace that file or adjust the script if you prefer a different source SVG.
