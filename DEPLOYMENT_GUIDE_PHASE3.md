# Phase 3 Deployment Guide

**Quick Reference for Deployment Team**

## What's Being Deployed

Unread discussion count badges for the Atoman Music Wiki system. Users will see the number of unread discussions directly on album cards and detail pages.

## Deployment Checklist

### Pre-Deployment (5 min)
- [ ] Pull latest code: `git pull origin main`
- [ ] Verify commits: `git log --oneline -2` (should show d7e8c9a and 605fc23)
- [ ] Check uncommitted changes: `git status` (should only show untracked docs)

### Build Phase (10 min)

**Backend:**
```bash
cd server
go build ./...
echo $?  # Should print 0 (success)
```

**Frontend:**
```bash
cd web
npm run build
# Check for build errors (should be none)
```

### Deployment (5-10 min)
1. Stop current services gracefully
2. Deploy new backend binary
3. Deploy new frontend build
4. No database migrations needed
5. No configuration changes needed
6. Restart services

### Post-Deployment Verification (10 min)

1. **Check Backend Health**
   ```bash
   curl http://localhost:8080/api/albums/{album-id}/discussions/unread-count
   # Should return: {"data": {"unread_count": N}}
   ```

2. **Check Frontend**
   - Navigate to music timeline
   - Blue badges should appear on albums with unread discussions
   - Click album detail page
   - Badge should appear in wiki section
   - Test on mobile device

3. **Monitor Logs**
   ```bash
   tail -f logs/server.log | grep -i "discussion\|error"
   # Should show no errors related to discussion endpoints
   ```

4. **Performance Check**
   - Response time for `/albums/*/discussions/unread-count` should be <100ms
   - No memory leaks (check process memory stable)
   - No database query errors

## Rollback Plan (If Needed)

If issues occur after deployment:

1. **Revert Commits**
   ```bash
   git revert d7e8c9a
   git revert 605fc23
   ```

2. **Rebuild and Redeploy**
   ```bash
   go build ./...
   npm run build
   # Deploy old version
   ```

3. **Investigation**
   - Check API endpoint availability
   - Verify database has `read_at` field on discussions table
   - Check browser console for JavaScript errors
   - Review application logs for any errors

## Key Changes Summary

| Component | Change | Impact | Risk |
|-----------|--------|--------|------|
| Backend | +2 API endpoints | Read-only queries | Low |
| Frontend | +Discussion badges | UI display only | Low |
| Database | +1 field on Discussion | Already exists in model | Low |

## Support Contact

For deployment issues:
1. Check logs first
2. Verify API endpoints are accessible
3. Test on both desktop and mobile
4. Contact backend team if API issues
5. Contact frontend team if UI issues

## Success Criteria

- ✅ No errors in logs post-deployment
- ✅ Discussion badges visible on music timeline
- ✅ Badges work on mobile
- ✅ API endpoints respond <100ms
- ✅ No console errors in browser DevTools

**Estimated Deployment Time:** 20-30 minutes
**Downtime Required:** < 5 minutes (graceful restart)
**Rollback Time:** < 10 minutes (if needed)
