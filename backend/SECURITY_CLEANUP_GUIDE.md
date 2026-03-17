# Security Cleanup Guide - CRITICAL P0

## Issue: .env File Committed to Git

The `.env` file containing sensitive configuration and weak placeholder secrets was committed to git history. This is a **CRITICAL** security vulnerability that must be addressed immediately.

## Impact

- JWT secrets exposed (even if placeholder values)
- Database credentials visible in git history
- AWS credentials structure revealed
- Security through obscurity compromised

## Step 1: Remove .env from Git History

**WARNING**: This rewrites git history. Coordinate with all team members before proceeding.

```bash
cd /Users/felixgeelhaar/Developer/projects/ascend

# Option A: Using git filter-branch (traditional method)
git filter-branch --force --index-filter \
  "git rm --cached --ignore-unmatch backend/.env" \
  --prune-empty --tag-name-filter cat -- --all

# Option B: Using BFG Repo-Cleaner (faster, recommended)
# Install BFG: brew install bfg
bfg --delete-files .env
git reflog expire --expire=now --all
git gc --prune=now --aggressive

# Step 2: Force push to all remotes
git push origin --force --all
git push origin --force --tags
```

## Step 2: Generate Strong Secrets

```bash
# Generate JWT Access Secret (32+ characters)
openssl rand -base64 32
# Example output: 7YvPC9oDxN2vCGGv3kF7Qq1nLZKxN8Vm9QKvR+sE4Xo=

# Generate JWT Refresh Secret (different from access)
openssl rand -base64 32
# Example output: mK3pL8nM5xR9vT2wY6qA4zH7jN0pS5uV8bC1eF4gI+D=

# Generate database password
openssl rand -base64 24
# Example output: kL9mN3pR7tV2xY5zA8cE4fH6j
```

## Step 3: Update .env.example

The `.env.example` file should contain:
- **Documentation**: Comments explaining each variable
- **Placeholder values**: Clearly fake values (e.g., `your_secret_here`)
- **Generation instructions**: How to generate secure values
- **NO real secrets**: Never commit actual credentials

```bash
# Update .env.example with proper placeholders
# See updated .env.example file
```

## Step 4: Verify .gitignore

Ensure `.gitignore` properly excludes environment files:

```gitignore
# Environment files
.env
.env.local
.env.*.local
.env.production
.env.staging
.env.development

# Keep .env.example tracked
!.env.example
```

## Step 5: Rotate All Secrets

After removing from git history, **rotate all secrets immediately**:

1. **JWT Secrets**: Generate new secrets as shown above
2. **Database Password**: Update PostgreSQL password
3. **AWS Credentials**: Rotate IAM access keys in AWS Console
4. **Redis Password**: Set a new Redis password (currently empty)
5. **Sentry DSN**: Regenerate Sentry DSN if configured
6. **PostHog API Key**: Rotate API key if configured

## Step 6: Set Up Secret Management

### For Development
Create your local `.env` file:

```bash
cd backend
cp .env.example .env
# Edit .env with your real secrets
nano .env
```

### For Production
Use environment-based secret management:

1. **AWS Secrets Manager** (recommended for production)
2. **HashiCorp Vault** (for multi-cloud)
3. **Kubernetes Secrets** (if using k8s)
4. **GitHub Actions Secrets** (for CI/CD)

Example AWS Secrets Manager setup:

```bash
# Store JWT secrets
aws secretsmanager create-secret \
  --name ascend/production/jwt-access-secret \
  --secret-string "$(openssl rand -base64 32)" \
  --region eu-west-1

aws secretsmanager create-secret \
  --name ascend/production/jwt-refresh-secret \
  --secret-string "$(openssl rand -base64 32)" \
  --region eu-west-1

# Retrieve in application
aws secretsmanager get-secret-value \
  --secret-id ascend/production/jwt-access-secret \
  --query SecretString \
  --output text
```

## Step 7: Update Deployment Documentation

Update deployment docs to include secret management:

```markdown
## Production Deployment Checklist

- [ ] Generate strong JWT secrets (32+ characters)
- [ ] Set up AWS Secrets Manager
- [ ] Configure environment variables in deployment platform
- [ ] Verify .env is NOT in Docker image
- [ ] Test secret rotation procedures
- [ ] Document secret access procedures for team
```

## Step 8: Team Communication

Send to all team members:

```
URGENT: .env File Security Issue

We've identified that .env was committed to git with placeholder secrets.

Actions Required:
1. Pull latest changes after history rewrite
2. Fetch all branches: git fetch origin --force
3. Reset your local branches: git reset --hard origin/main
4. Create new .env from .env.example
5. Generate strong secrets using: openssl rand -base64 32
6. NEVER commit .env file

Questions? Contact: [security-contact]
```

## Step 9: Implement Pre-commit Hook

Prevent future accidents with a pre-commit hook:

```bash
# .git/hooks/pre-commit
#!/bin/bash

# Check if .env is being committed
if git diff --cached --name-only | grep -q "\.env$"; then
    echo "ERROR: Attempting to commit .env file!"
    echo "This file contains secrets and should never be committed."
    echo "Please remove it from staging: git reset HEAD .env"
    exit 1
fi

# Check for potential secrets in committed files
git diff --cached | grep -E "(password|secret|key|token|api_key)" && \
    echo "WARNING: Potential secret detected in commit. Please review carefully."

exit 0
```

Make it executable:

```bash
chmod +x .git/hooks/pre-commit
```

## Step 10: Security Audit Log

Document this incident:

```yaml
incident_date: 2025-01-XX
severity: CRITICAL
issue: .env file with placeholder secrets committed to git
impact: Low (placeholder values only, no production exposure)
resolution_steps:
  - Removed from git history
  - Updated .env.example with documentation
  - Verified .gitignore configuration
  - Added pre-commit hook
  - Updated team documentation
lessons_learned:
  - Implement pre-commit hooks early in project
  - Conduct regular security reviews
  - Use secret scanning tools (e.g., TruffleHog, git-secrets)
```

## Verification Checklist

- [ ] .env removed from git history
- [ ] Force push completed to all remotes
- [ ] All team members notified and pulled changes
- [ ] New secrets generated
- [ ] .env.example updated with proper documentation
- [ ] .gitignore verified
- [ ] Pre-commit hook installed
- [ ] Secret management strategy documented
- [ ] Deployment guides updated
- [ ] Security incident logged

## Prevention for Future

1. **Use secret scanning tools**:
   ```bash
   # Install git-secrets
   brew install git-secrets
   cd /path/to/repo
   git secrets --install
   git secrets --register-aws
   ```

2. **GitHub Advanced Security**: Enable secret scanning in repository settings

3. **CI/CD secret scanning**: Integrate TruffleHog or GitGuardian

4. **Regular security reviews**: Weekly audits of committed files

5. **Developer training**: Security awareness sessions on secret management

## Resources

- [OWASP Secret Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [AWS Secrets Manager Best Practices](https://docs.aws.amazon.com/secretsmanager/latest/userguide/best-practices.html)
- [GitHub Security Best Practices](https://docs.github.com/en/code-security/getting-started/best-practices-for-preventing-data-leaks-in-your-organization)

---

**STATUS**: This guide must be executed immediately before any production deployment.
