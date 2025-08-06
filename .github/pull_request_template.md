# Pull Request

## Summary

**Brief description of changes:**
<!-- Provide a clear and concise summary of what this PR accomplishes -->

**Related Issue(s):**
<!-- Link related issues: Fixes #123, Closes #456, Related to #789 -->

## Type of Change

**Select the type of change this PR introduces:**
- [ ] **Bug fix** (non-breaking change that fixes an issue)
- [ ] **New feature** (non-breaking change that adds functionality)
- [ ] **Breaking change** (fix or feature that causes existing functionality to not work as expected)
- [ ] **Documentation update** (changes to documentation only)
- [ ] **Refactoring** (code changes that neither fix bugs nor add features)
- [ ] **Performance improvement** (changes that improve performance)
- [ ] **Test improvement** (adding missing tests or correcting existing tests)
- [ ] **Build/CI** (changes to build process or continuous integration)
- [ ] **Other** (specify): ___________

## Changes Made

**Detailed description of changes:**
<!-- Explain the changes in detail. Include motivation and implementation approach. -->

### Modified Components
- [ ] Controllers
- [ ] API definitions (CRDs)
- [ ] Database utilities
- [ ] Configuration
- [ ] Documentation
- [ ] Tests
- [ ] CI/CD
- [ ] Examples
- [ ] Other: ___________

### Database Support
<!-- If this change affects specific database types, check all that apply -->
- [ ] MySQL Replication
- [ ] MySQL Group Replication  
- [ ] Redis Replication
- [ ] Redis Cluster
- [ ] PostgreSQL Replication
- [ ] ProxySQL Sync
- [ ] Generic/All databases

## Breaking Changes

**Are there any breaking changes:**
- [ ] No breaking changes
- [ ] Yes, breaking changes (describe below)

**If yes, describe the breaking changes and migration path:**
<!-- Provide detailed information about what breaks and how users can migrate -->

```yaml
# Example of old configuration
---
# Example of new configuration
```

## Testing

**Testing performed:**
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] End-to-end tests added/updated
- [ ] Manual testing performed
- [ ] No testing required

**Test coverage:**
- [ ] Coverage increased or maintained
- [ ] Coverage decreased (explain why): ___________

**Manual testing details:**
```bash
# Describe the manual testing steps performed
# Include commands run and expected vs actual results
```

**Test environments:**
- [ ] Local development environment
- [ ] CI/CD pipeline
- [ ] Staging/test cluster
- [ ] Production-like environment

## Documentation

**Documentation changes:**
- [ ] No documentation changes needed
- [ ] Documentation updated in this PR
- [ ] Documentation will be updated in a separate PR
- [ ] New documentation created

**If documentation updated, what was changed:**
- [ ] User guide
- [ ] API reference
- [ ] Examples
- [ ] Architecture diagrams
- [ ] Troubleshooting guide
- [ ] Developer guide
- [ ] README
- [ ] Code comments
- [ ] Other: ___________

## Security Considerations

**Security review needed:**
- [ ] No security implications
- [ ] Minor security implications (explain below)
- [ ] Significant security implications (detailed review required)

**Security considerations:**
<!-- If there are security implications, describe them -->

## Performance Impact

**Performance impact assessment:**
- [ ] No performance impact
- [ ] Positive performance impact
- [ ] Negative performance impact (justify below)
- [ ] Unknown performance impact

**Performance details:**
<!-- If there's a performance impact, provide details -->

## Deployment Considerations

**Deployment requirements:**
- [ ] No special deployment requirements
- [ ] Requires operator restart
- [ ] Requires CRD updates
- [ ] Requires database configuration changes
- [ ] Requires Kubernetes version upgrade
- [ ] Other requirements: ___________

**Rollback considerations:**
- [ ] Changes are fully reversible
- [ ] Rollback requires special steps (describe below)
- [ ] Rollback not possible without data loss

## Checklist

**Before requesting review:**
- [ ] I have read the [contributing guidelines](../CONTRIBUTING.md)
- [ ] I have followed the [code style guidelines](../CONTRIBUTING.md#coding-standards)
- [ ] I have performed a self-review of my code
- [ ] I have commented my code where necessary
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published

**Code quality:**
- [ ] Code follows project conventions
- [ ] Error handling is appropriate
- [ ] Logging is appropriate and follows project standards
- [ ] Code is properly commented
- [ ] No debugging code left in

**Testing quality:**
- [ ] Tests cover the new functionality
- [ ] Tests include edge cases
- [ ] Tests are maintainable and readable
- [ ] Mock dependencies are used appropriately

## Screenshots

**If this PR includes UI changes or visual examples:**
<!-- Drag and drop images here or link to them -->

**Before:**
<!-- Screenshot or description of the previous behavior -->

**After:**
<!-- Screenshot or description of the new behavior -->

## Additional Notes

**Reviewer focus areas:**
<!-- Highlight specific areas where you'd like reviewers to focus -->

**Known limitations:**
<!-- Any known limitations or future improvements needed -->

**Dependencies:**
<!-- Any dependencies on other PRs, issues, or external changes -->

---

**Additional Context:**
<!-- Any other information that would be helpful for reviewers -->