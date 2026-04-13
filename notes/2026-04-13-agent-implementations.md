# Agent implementations

## Abstract

Agents are a form of engineering around the flaky, yet versatile nature of
large language models (LLM). With tool use, an LLM recognizes a data
acquisition need and will ask the environment to provide the data, if a
suitable tool is available.

The agent then is a composition of those tools and the loop. There is a lot of
flexibility in how you want to structure that. There is also a lot of value in
having datastores at hand, that would be able to respond to queries. One
perspective on that is that every company that wants to remain visible, may
want to expose what the do or provide through an API. A website is for humans,
but a machine would require something else.

