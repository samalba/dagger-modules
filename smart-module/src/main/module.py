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

from . import tools


@object_type
class SmartModule:
    def _init_tools(self) -> list:
        """Initialize the tools for the LLM"""
        tools_funcs = []
        for t in dir(tools):
            if not t.startswith("tool_"):
                continue
            func = getattr(tools, t)
            if not callable(func):
                continue
            tools_funcs.append(getattr(tools, t))
        return tools_funcs

    @function
    async def ask(self, api_key: dagger.Secret, prompt: str) -> str:
        """Ask the LLM a prompt that involves a dagger module call"""
        llm = ChatOpenAI(model="gpt-4o", api_key=await api_key.plaintext())

        tools = self._init_tools()
        llm_with_tools = llm.bind_tools(tools)

        builder = StateGraph(MessagesState)

        async def assistant(state: MessagesState):
            sys_msg = SystemMessage(content="You are a helpful assistant tasked with performing arithmetic on a set of inputs.")
            response = await llm_with_tools.ainvoke([sys_msg])
            return {"messages": [response + state["messages"]]}

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

        inputs = {"messages": [HumanMessage(content=prompt)]}
        messages = await graph.ainvoke(inputs)

        for m in messages['messages']:
            m.pretty_print()

        return messages['messages'][-1].content
