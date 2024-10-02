from dagger import dag
import dagger

from pydantic import BaseModel


class Directory(BaseModel):
    d_directory_id: str

class Container(BaseModel):
    d_container_id: str


async def t_container_from_image_id(image: str) -> Container:
    """Create a Container from an image.
    Calls this to create a Container from an image id.

    Args:
        image: image id.

    Returns:
        The Container object.
    """

    ctr = dag.container().from_(image)
    return Container(d_container_id=await ctr.id())

async def t_container_from_directory(directory: Directory) -> Container:
    """Create a Container from a Directory.
    Calls this to build a Container from a Directory.

    Args:
        directory: the Directory object

    Returns:
        The Container object.
    """

    #FIXME: support directory type (currently limited to building from the root dir)
    dir_id = dagger.DirectoryID(directory.d_directory_id)
    container = dag.container().build(context=dag.load_directory_from_id(dir_id))
    return Container(d_container_id=await container.id())

async def t_git_pull(repository_url: str, branch: str) -> Directory:
    """Pulls a git repository.
    Calls this to pull code from a git repository and Directory.

    Args:
        repository_url: git repository URL name.
        branch: branch to pull.

    Returns:
        The Directory object.
    """

    directory = await dag.git(repository_url).branch(branch).tree()
    return Directory(d_directory_id=await directory.id())

async def t_container_publish(container: Container, address: str) -> str:
    """Publish a Container to a registry.
    Calls this to publish or push a Container to a registry address.

    Args:
        container: Container object
        address: full url of the remote image, including the registry

    Returns:
        The address of the published container.
    """
    ctr_id = dagger.ContainerID(container.d_container_id)
    return await dag.load_container_from_id(ctr_id).publish(address=address)
