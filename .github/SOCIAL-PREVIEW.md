Repository social preview

This repository includes a social preview SVG that can be exported to PNG for GitHub's social preview image.

Files:
- `.github/social-preview.svg` — the source SVG (if present).
- `assets/brand/favicons/social-preview.png` — recommended output path for export.

How to set the social preview on GitHub:

1. Generate `social-preview.png` with:

```bash
bash scripts/generate-favicons.sh
```

2. In GitHub: Repository -> Settings -> Social preview -> Upload the `assets/brand/favicons/social-preview.png` file.

Automating via `gh` (GitHub CLI):

Uploading via API requires authentication and calling the repository settings API. It's simpler to upload via the web UI. If you want to automate, use `gh api` with GitHub's REST API and a personal access token with `repo` scope.
