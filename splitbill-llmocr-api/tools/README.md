# Tools

This directory contains utility tools for development and debugging.

## Database URL Tester

The `test_db_url.go` tool helps validate your `DATABASE_URL` environment variable before running the main application.

### Usage

1. **Set your DATABASE_URL environment variable:**
   ```bash
   export DATABASE_URL="your_supabase_connection_string"
   ```

2. **Run the test tool:**
   ```bash
   cd tools
   go run test_db_url.go
   ```

3. **Check the output:**
   - ✅ Green checkmarks indicate valid components
   - ❌ Red X marks indicate errors that need fixing
   - ⚠️ Yellow warnings indicate potential issues

### Example Output

```
🔍 Testing DATABASE_URL: postgresql://postgres.***@aws-0-us-east-1.pooler.supabase.com:6543/postgres?sslmode=require
✅ Hostname: aws-0-us-east-1.pooler.supabase.com
✅ Port: 6543
✅ Username: postgres
✅ Password: [present]
✅ Database: postgres
✅ SSL Mode: require

✅ DATABASE_URL is valid!
```

### Common Issues

1. **Missing protocol prefix** - Must start with `postgresql://` or `postgres://`
2. **Missing hostname** - Check your Supabase project settings
3. **Missing username/password** - Verify your database credentials
4. **Missing database name** - Usually `postgres` for Supabase
5. **Missing SSL mode** - Supabase requires `sslmode=require`

### Supabase Connection String Format

```
postgresql://postgres.[project-ref]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres?sslmode=require
```

Where:
- `[project-ref]` is your Supabase project reference ID
- `[password]` is your database password
- `[region]` is your Supabase project region
