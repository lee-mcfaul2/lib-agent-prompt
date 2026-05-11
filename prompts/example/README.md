# Example prompt

This directory contains a single example prompt for the demo bundle.

The `services[].schema_digest` and `service-references/postgresql-service.json:source_digest` fields use placeholder all-zero digests because `agent-sql-mcp` has not yet produced its first signed `mcp-schema.json`. Real bundle builds against a live `agent-sql-mcp` substitute the actual digest at build time.
