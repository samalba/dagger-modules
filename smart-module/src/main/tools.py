from dagger import dag

# from langchain_core.tools import tool

# @tool
async def tool_build_and_push(repository_url: str, branch: str, ref: str) -> str:
    """Builds and pushes a Docker image from a git repository to a registry.

    Args:
        repository_url: git repository URL name
        branch: branch to build from
        ref: full url of the remote image, including the registry
    """

    dir = dag.git(repository_url).branch(branch).tree()
    return await dag.container().build(context=dir).publish(address=ref)
