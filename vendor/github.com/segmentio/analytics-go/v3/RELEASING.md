# Releasing

1. **Update the version constant** in `analytics.go`:
   ```go
   const Version = "x.y.z"
   ```

2. **Update `History.md`** with a summary of changes and the release date.

3. Commit both files:
   ```
   git commit -m "Release vx.y.z"
   ```

4. Push and merge to `v3.0`.

5. Tag the release from `v3.0` after the merge:
   ```
   git tag vx.y.z
   git push origin vx.y.z
   ```

> **Note:** The `Version` constant in `analytics.go` is the single source of truth — it is sent in the `User-Agent` header on every API request and is automatically used by the test fixtures. No other files need to be updated.
