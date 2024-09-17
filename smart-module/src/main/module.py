"""A generated module for SmartModule functions

This module has been generated via dagger init and serves as a reference to
basic module structure as you get started with Dagger.

Two functions have been pre-created. You can modify, delete, or add to them,
as needed. They demonstrate usage of arguments and return types using simple
echo and grep commands. The functions can be called from the dagger CLI or
from one of the SDKs.

The first line in this comment block is a short description line and the
rest is a long description with more detail on the module's purpose or usage,
if appropriate. All modules should have a short description.
"""

import dagger
from dagger import dag, function, object_type

from langchain_core.messages import HumanMessage, SystemMessage
from langchain_openai import ChatOpenAI
from langgraph.graph import START, StateGraph, MessagesState
from langgraph.prebuilt import tools_condition, ToolNode

from .tools import multiply


@object_type
class SmartModule:
    def _init_tools(self) -> list:
        """Initialize the tools for the LLM"""
        tools = []
        tools.append(multiply)
        return tools

    @function
    async def ask(self, api_key: dagger.Secret, prompt: str) -> str:
        """Ask the LLM a prompt that involves a dagger module call"""
        llm = ChatOpenAI(model="gpt-4o", api_key=await api_key.plaintext())

        tools = self._init_tools()
        llm_with_tools = llm.bind_tools(tools)

        builder = StateGraph(MessagesState)

        def assistant(state: MessagesState):
            sys_msg = SystemMessage(content="You are a helpful assistant tasked with performing arithmetic on a set of inputs.")
            return {"messages": [llm_with_tools.invoke([sys_msg] + state["messages"])]}

        # Define nodes: these do the work
        builder.add_node("assistant", assistant)
        builder.add_node("tools", ToolNode(tools))

        # Define edges: these determine how the control flow moves
        builder.add_edge(START, "assistant")
        builder.add_conditional_edges(
            "assistant",
            # If the latest message (result) from assistant is a tool call -> tools_condition routes to tools
            # If the latest message (result) from assistant is a not a tool call -> tools_condition routes to END
            tools_condition,
        )
        builder.add_edge("tools", "assistant")
        graph = builder.compile()

        messages = [HumanMessage(content=prompt)]
        messages = graph.invoke({"messages": messages})

        for m in messages['messages']:
            m.pretty_print()

        return messages['messages'][-1].content


    # @function
    # async def grep_dir(self, directory_arg: dagger.Directory, pattern: str) -> str:
    #     """Returns lines that match a pattern in the files of the provided Directory"""
    #     return await (
    #         dag.container()
    #         .from_("alpine:latest")
    #         .with_mounted_directory("/mnt", directory_arg)
    #         .with_workdir("/mnt")
    #         .with_exec(["grep", "-R", pattern, "."])
    #         .stdout()
    #     )
