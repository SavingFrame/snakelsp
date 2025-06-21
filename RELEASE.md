# Release Process

This document describes how to create releases for SnakeLSP.

## Automated Releases

The project uses GitHub Actions to automatically create releases when you push a git tag.

### Creating a Release

1. **Create and push a tag:**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **The workflow will automatically:**
   - Run all tests
   - Build binaries for multiple platforms:
     - Linux (AMD64, ARM64)
     - macOS (AMD64, ARM64)
   - Generate a changelog from git commits
   - Create a GitHub release with all binaries attached

### Tag Naming Convention

Use semantic versioning with a `v` prefix:
- `v1.0.0` - Major release
- `v1.1.0` - Minor release  
- `v1.1.1` - Patch release
- `v1.0.0-beta.1` - Pre-release

### Changelog Generation

The changelog is automatically generated from git commit messages between the current and previous tag. To get better changelogs:

- Use descriptive commit messages
- Follow conventional commits format (optional but recommended):
  - `feat: add new feature`
  - `fix: resolve bug`
  - `docs: update documentation`
  - `refactor: improve code structure`

### Manual Release (if needed)

If you need to create a release manually:

1. Go to GitHub → Releases → "Create a new release"
2. Choose or create a tag
3. Fill in the release notes
4. Upload binaries manually (not recommended)

### Troubleshooting

- **Workflow fails**: Check the Actions tab for error details
- **Missing binaries**: Ensure the build step completed successfully
- **Permission errors**: Verify the repository has proper permissions for GitHub Actions
