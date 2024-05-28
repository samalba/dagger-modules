"""Dagger module for interacting with Snowflake via the Snowflake CLI.

This module provides a `SnowflakeCli` object type that can be used to run Snowflake CLI commands.
"""

import uuid

import dagger
from dagger import dag, function, object_type


"""Dagger module for interacting with Snowflake via the Snowflake CLI."""
@object_type
class SnowflakeCli:
    @function
    def ctr(self) -> dagger.Container:
        """Run a Snowflake CLI command"""
        return (
            dag.container().from_("python:3-alpine")
            .with_exec(["apk", "add", "--no-cache", "alpine-sdk", "libffi-dev"])
            .with_exec(["pip", "install", "snowflake-cli-labs"])
        )

    """Run a SQL query using the Snowflake CLI command"""
    @function
    def query(self, config: dagger.Secret, query: str, format_json: bool = False, cache: bool = True) -> str:
        """Run a query"""
        args = ["snow", "--config-file", "/config.toml", "sql", "-i"]
        if format_json:
            args.append("--format=json")
        ctr = self.ctr()
        if cache is False:
            ctr = ctr.with_env_variable("_SNOWFLAKE_QUERY_CACHE_BUSTER", uuid.uuid4().hex)
        ctr = ctr.with_mounted_secret("/config.toml", config)
        return ctr.with_exec(args, stdin=query).stdout()
