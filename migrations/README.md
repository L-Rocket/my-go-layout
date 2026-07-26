# migrations

SQL migrations for the PostgREST/Supabase storage backend.

Naming convention: timestamped versions — `YYYYMMDDHHMMSS_name.sql`.
Plain counters (`0001`, `0002`, …) collide when several repos push
migrations into one shared Supabase project; timestamps don't.
