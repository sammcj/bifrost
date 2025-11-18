# Helm Chart GitHub Pages Setup Guide

## Current Issue

ArgoCD is getting 404 when trying to fetch: `https://maximhq.github.io/bifrost/helm-charts/index.yaml`

The `gh-pages` branch exists but the files may not be deployed correctly or GitHub Pages isn't configured.

## Steps to Fix

### 1. Check GitHub Pages Configuration

Go to: https://github.com/maximhq/bifrost/settings/pages

Ensure:
- **Source**: Deploy from a branch
- **Branch**: `gh-pages`
- **Folder**: `/ (root)`
- Click **Save** if any changes were made

### 2. Verify gh-pages Branch Content

The `gh-pages` branch should have this structure:
```
helm-charts/
├── index.yaml
└── bifrost-1.3.5.tgz
```

Check the current content:
```bash
git fetch origin gh-pages
git checkout gh-pages
ls -la helm-charts/
```

### 3. Manually Trigger the Workflow

Since the workflow already exists, trigger it to deploy properly:

1. Go to: https://github.com/maximhq/bifrost/actions/workflows/helm-release.yml
2. Click **Run workflow**
3. Select `main` branch
4. Click **Run workflow**

Or via command line:
```bash
# From the main branch
git checkout main
git commit --allow-empty -m "chore: trigger helm chart deployment"
git push origin main
```

### 4. Verify Workflow Permissions

Go to: https://github.com/maximhq/bifrost/settings/actions

Under **Workflow permissions**:
- Select: **Read and write permissions** ✓
- Check: **Allow GitHub Actions to create and approve pull requests** ✓
- Click **Save**

### 5. Wait for Deployment

After the workflow runs successfully:
- Wait 2-3 minutes for GitHub Pages to propagate
- Check the Actions tab for the workflow status
- Look for the "Deploy to GitHub Pages" step completion

### 6. Test the Deployment

```bash
# Test if index is accessible
curl -I https://maximhq.github.io/bifrost/helm-charts/index.yaml

# Should return: HTTP/2 200

# Test full index content
curl https://maximhq.github.io/bifrost/helm-charts/index.yaml

# Should return YAML content with chart metadata
```

### 7. Test Helm Repository

```bash
# Add the repository
helm repo add bifrost https://maximhq.github.io/bifrost/helm-charts
helm repo update

# Search for charts
helm search repo bifrost

# Should show:
# NAME              CHART VERSION  APP VERSION  DESCRIPTION
# bifrost/bifrost   1.3.5          1.3.5        A Helm chart for deploying Bifrost...
```

## Troubleshooting

### If Still Getting 404

1. **Check gh-pages branch directly**:
   - Visit: https://github.com/maximhq/bifrost/tree/gh-pages
   - Verify `helm-charts/index.yaml` file exists
   - Click on the file to view its content

2. **Check GitHub Pages URL**:
   - The base URL should be: `https://maximhq.github.io/bifrost/`
   - The chart repo URL is: `https://maximhq.github.io/bifrost/helm-charts/`
   - Visit: https://maximhq.github.io/bifrost/helm-charts/ in browser

3. **Check Workflow Logs**:
   - Go to Actions tab
   - Click on the most recent "Release Helm Chart" workflow
   - Check each step for errors
   - Verify "Deploy to GitHub Pages" step succeeded

4. **Force Re-deployment**:
   ```bash
   # Delete and recreate gh-pages branch
   git push origin --delete gh-pages
   
   # Trigger workflow to recreate it
   git commit --allow-empty -m "chore: recreate gh-pages"
   git push origin main
   ```

### If Workflow Fails

Check for these common issues:

1. **Permission Error**: 
   - Fix: Enable write permissions in Settings → Actions → General

2. **Branch Protection**:
   - Fix: Ensure `gh-pages` branch has no protection rules that block Actions

3. **Pages Not Enabled**:
   - Fix: Enable in Settings → Pages

### Verify Current State

Run these commands to check everything:

```bash
# Check if Pages is accessible
curl -I https://maximhq.github.io/bifrost/

# Check if helm-charts subdirectory is accessible
curl -I https://maximhq.github.io/bifrost/helm-charts/

# Check if index.yaml exists
curl -I https://maximhq.github.io/bifrost/helm-charts/index.yaml

# Check if chart package exists
curl -I https://maximhq.github.io/bifrost/helm-charts/bifrost-1.3.5.tgz
```

All should return `HTTP/2 200`.

## Expected Workflow Output

When the workflow runs successfully, you should see:

1. ✓ Checkout
2. ✓ Configure Git  
3. ✓ Install Helm
4. ✓ Run chart-testing (lint)
5. ✓ Package Helm chart
6. ✓ Upload artifacts
7. ✓ Deploy to GitHub Pages

The last step should show:
```
Publishing to gh-pages
✓ Successfully published
```

## For ArgoCD Integration

Once the Helm repository is accessible, your ArgoCD/Kustomize should work with:

```yaml
helmCharts:
  - name: bifrost
    repo: https://maximhq.github.io/bifrost/helm-charts
    version: "1.3.5"
    releaseName: bifrost
```

## Contact

If issues persist after following these steps, check:
- GitHub Actions logs: https://github.com/maximhq/bifrost/actions
- GitHub Pages status: https://github.com/maximhq/bifrost/settings/pages
- Workflow file: `.github/workflows/helm-release.yml`
