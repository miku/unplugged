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

## Zoo of Agents (cli)

Endless list
[e2b-dev/awesome-ai-agents](https://github.com/e2b-dev/awesome-ai-agents) with
20k stars, not updated in a year? Some repos have 10k+ stars, but seem
abandonned now. Cambrian explosion, lots of life, some of it short-lived;
[awesome-code-ai](https://github.com/sourcegraph/awesome-code-ai) "was archived
by the owner on Feb 23, 2026" - shorter, maybe incomplete list;
[more](https://github.com/ikaijua/Awesome-AITools?tab=readme-ov-file#ai-agent)
and more and more, ...

Some more, that I had a chance to run:

* [crush](https://github.com/charmbracelet/crush)
* [goose](https://goose-docs.ai/)
* [claude code](https://en.wikipedia.org/wiki/Claude_(language_model)#Claude_Code), [2025-02-24](https://www.anthropic.com/news/claude-3-7-sonnet), [best practices](https://code.claude.com/docs/en/best-practices), ...
* [pi](https://pi.dev)

**So, when everybody does mostly the same, but slightly different, what could be
a better design?**

Maybe to keep the default experience minimal, but the program extensible?

Also, I notice a large number of additional tools, that may come in handy, but
you do not want to keep N tools around, if you could somehow use a single one,
that has the option for plugins or the like.

Example: Automatic stdout/stderr compaction or common commands, etc.

## a74c5da1 2025-08-09 Initial monorepo setup ...

![](static/git-cal-pi-mono.png)

As of 04/2026, it consists of 211450 LOC (ts: 167246, md: 17817, ...); 2988
directories, 33629 files; with 80% of the code committed by the original
author.

```shell
$ git summary | head -20

 project     : pi-mono
 repo age    : 8 months
 branch:     : main
 last active : 18 hours ago
 active on   : 184 days
 commits     : 3515
 files       : 775
 uncommitted : 0
 authors     :
  2765  Mario Zechner                 78.7%
    66  Helmut Januschka              1.9%
    59  github-actions[bot]           1.7%
    42  Aliou Diallo                  1.2%
    41  Markus Ylisiurunen            1.2%
    36  Sviatoslav Abakumov           1.0%
    29  Nico Bailon                   0.8%
    22  Armin Ronacher                0.6%
    18  Danila Poyarkov               0.5%
    18  mom                           0.5%
```

## Layered architecture

* ai/llm layer (abstractions over various providers)
* agent (generic agent interaction)
* tui
* coding-agent
* ...

Minimal tool support:

```
Available Tools (default: read, bash, edit, write):
  read   - Read file contents
  bash   - Execute bash commands
  edit   - Edit files with find/replace
  write  - Write files (creates/overwrites)
  grep   - Search file contents (read-only, off by default)
  find   - Find files by glob pattern (read-only, off by default)
  ls     - List directory contents (read-only, off by default)
```

Basically betting that bash is a concise way to access information.

On start, no plugins:

```shell
$ pi list
No packages installed.
```

## The familiar loop

Lifecycle overview:

```
pi starts
  │
  ├─► session_start { reason: "startup" }
  └─► resources_discover { reason: "startup" }
      │
      ▼
user sends prompt ─────────────────────────────────────────┐
  │                                                        │
  ├─► (extension commands checked first, bypass if found)  │
  ├─► input (can intercept, transform, or handle)          │
  ├─► (skill/template expansion if not handled)            │
  ├─► before_agent_start (can inject message, modify system prompt)
  ├─► agent_start                                          │
  ├─► message_start / message_update / message_end         │
  │                                                        │
  │   ┌─── turn (repeats while LLM calls tools) ───┐       │
  │   │                                            │       │
  │   ├─► turn_start                               │       │
  │   ├─► context (can modify messages)            │       │
  │   ├─► before_provider_request (can inspect or replace payload)
  │   │                                            │       │
  │   │   LLM responds, may call tools:            │       │
  │   │     ├─► tool_execution_start               │       │
  │   │     ├─► tool_call (can block)              │       │
  │   │     ├─► tool_execution_update              │       │
  │   │     ├─► tool_result (can modify)           │       │
  │   │     └─► tool_execution_end                 │       │
  │   │                                            │       │
  │   └─► turn_end                                 │       │
  │                                                        │
  └─► agent_end                                            │
                                                           │
user sends another prompt ◄────────────────────────────────┘

/new (new session) or /resume (switch session)
  ├─► session_before_switch (can cancel)
  ├─► session_shutdown
  ├─► session_start { reason: "new" | "resume", previousSessionFile? }
  └─► resources_discover { reason: "startup" }

/fork
  ├─► session_before_fork (can cancel)
  ├─► session_shutdown
  ├─► session_start { reason: "fork", previousSessionFile }
  └─► resources_discover { reason: "startup" }

/compact or auto-compaction
  ├─► session_before_compact (can cancel or customize)
  └─► session_compact

/tree navigation
  ├─► session_before_tree (can cancel or customize)
  └─► session_tree

/model or Ctrl+P (model selection/cycling)
  └─► model_select

exit (Ctrl+C, Ctrl+D, SIGHUP, SIGTERM)
  └─► session_shutdown
```

Plugins can hook into the various stages, e.g.

* `tool_result`, `resources_discover`

## Other notable features

* switch model in session
* lots of packages: [shittycodingagent.ai/packages](https://shittycodingagent.ai/packages), just be careful

## Examples


### Extension: Additional resources



### Extension: Web Access

```
$ pi install npm:pi-web-access
```

### Skills

## Other agent
