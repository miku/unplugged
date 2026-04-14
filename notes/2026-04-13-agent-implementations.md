# Agents implementations

> Martin Czygan, Lightning Talk at [KI MEETUP
> HALLE](https://www.meetup.com/de-de/webwirtschaft/events/314258534/)
> [#11](https://klaustor-coworking.de/events/260415-ki-meetup-halle/),
> 2026-04-15, [Klaustor Coworking](https://klaustor-coworking.de/)

## Intro

**Agents** are a form of **engineering** around the flaky, yet versatile nature
of large language models (LLM). With tool use, an LLM recognizes a data
acquisition need and will request additional input.

The agent then is a composition of those tools and the loop. There is a lot of
flexibility in how you want to structure that.

* [building effective agents](https://www.anthropic.com/engineering/building-effective-agents) (12/2024)

> Consistently, the most successful implementations use simple, composable
> patterns rather than complex frameworks.

There is the tension between simplicity and transparency and ease of implementation.

> These frameworks make it easy to get started by simplifying standard
> low-level tasks like calling LLMs, defining and parsing tools, and chaining
> calls together. However, they often create extra layers of abstraction that
> can obscure the underlying prompts and responses, making them harder to
> debug. They can also make it tempting to add complexity when a simpler setup
> would suffice.

We saw a raw "agent" (standard lib and HTTP only) that had access to bash in
[Writing a simple agent in Go](https://github.com/miku/unplugged) (cf. KI
MEETUP #9. ...) and could answer questions about the weather and help with some
simple OS questions.

## Benchmarks for tool use

* [evals for AI agents](https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents) (01/2026)


There is also a lot of value in having datastores at hand, that would be able
to respond to queries. One perspective on that is that every company that wants
to remain visible, may want to expose what the do or provide through an API. A
website is for humans, but a machine would require something else.

