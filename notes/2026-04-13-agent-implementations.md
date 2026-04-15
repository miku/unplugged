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
$ $ pi list
User packages:
  npm:pi-web-access
    /home/tir/.npm-packages/lib/node_modules/pi-web-access
```

![](static/pi-with-plugin.png)

Find installed package in settings:

```
$ jq . ~/.pi/agent/settings.json
{
  "lastChangelogVersion": "0.67.2",
  "defaultProvider": "ollama/strix",
  "defaultModel": "nemotron-3-super:latest",
  "packages": [
    "npm:pi-web-access"
  ]
}
```

### Skills

* [agentskills](https://agentskills.io/client-implementation/adding-skills-support) spec, included in prompt

```
<available_skills>
  <skill>
    <name>pdf-processing</name>
    <description>Extract PDF text, fill forms, merge files. Use when handling PDFs.</description>
    <location>/home/user/.agents/skills/pdf-processing/SKILL.md</location>
  </skill>
  <skill>
    <name>data-analysis</name>
    <description>Analyze datasets, generate charts, and create summary reports.</description>
    <location>/home/user/project/.agents/skills/data-analysis/SKILL.md</location>
  </skill>
</available_skills>
```

Example skill:

```
$ unzip -l simonlee2-ffmpeg-toolkit.zip
Archive:  simonlee2-ffmpeg-toolkit.zip
  Length      Date    Time    Name
---------  ---------- -----   ----
     8690  2026-04-15 16:22   ffmpeg-toolkit/SKILL.md
     8679  2026-04-15 16:22   ffmpeg-toolkit/references/ffmpeg-commands.md
---------                     -------
    17369                     2 files


```

Example file conversion.


## Other agents

* GLM analyzes code: [lgo-58/2026-04-01-CC-BY-GLM.md](https://github.com/golang-leipzig/golang-leipzig.github.io/blob/source/static/downloads/lgo-58/2026-04-01-CC-BY-GLM.md)

Some of the prompts:

```
$ fd 'prompt.ts'
src/buddy/prompt.ts
src/components/FeedbackSurvey/TranscriptSharePrompt.tsx
src/components/ShowInIDEPrompt.tsx
src/components/permissions/PermissionPrompt.tsx
src/services/autoDream/consolidationPrompt.ts
src/services/compact/prompt.ts
src/tools/AgentTool/prompt.ts
src/tools/AskUserQuestionTool/prompt.ts
src/tools/BashTool/prompt.ts
src/tools/BriefTool/prompt.ts
src/tools/ConfigTool/prompt.ts
src/tools/EnterPlanModeTool/prompt.ts
src/tools/EnterWorktreeTool/prompt.ts
src/tools/ExitPlanModeTool/prompt.ts
src/tools/ExitWorktreeTool/prompt.ts
src/tools/FileEditTool/prompt.ts
src/tools/FileReadTool/prompt.ts
src/tools/FileWriteTool/prompt.ts
src/tools/GlobTool/prompt.ts
src/tools/GrepTool/prompt.ts
src/tools/LSPTool/prompt.ts
src/tools/ListMcpResourcesTool/prompt.ts
src/tools/MCPTool/prompt.ts
src/tools/NotebookEditTool/prompt.ts
src/tools/PowerShellTool/prompt.ts
src/tools/ReadMcpResourceTool/prompt.ts
src/tools/RemoteTriggerTool/prompt.ts
src/tools/ScheduleCronTool/prompt.ts
src/tools/SendMessageTool/prompt.ts
src/tools/SkillTool/prompt.ts
src/tools/SleepTool/prompt.ts
src/tools/TaskCreateTool/prompt.ts
src/tools/TaskGetTool/prompt.ts
src/tools/TaskListTool/prompt.ts
src/tools/TaskStopTool/prompt.ts
src/tools/TaskUpdateTool/prompt.ts
src/tools/TeamCreateTool/prompt.ts
src/tools/TeamDeleteTool/prompt.ts
src/tools/TodoWriteTool/prompt.ts
src/tools/ToolSearchTool/prompt.ts
src/tools/WebFetchTool/prompt.ts
src/tools/WebSearchTool/prompt.ts
src/utils/claudeInChrome/prompt.ts
src/utils/processUserInput/processTextPrompt.ts
src/utils/swarm/It2SetupPrompt.tsx
src/utils/systemPrompt.ts
```

Still amazed be the STEERING BY UPPERCASE!

> Example take from the "src/tools/BashTool/prompt.ts"

Only create commits when requested by the user. If unclear, ask first. When the user asks you to create a new git commit, follow these steps carefully:

You can call multiple tools in a single response. When multiple independent pieces of information are requested and all commands are likely to succeed, run multiple tool calls in parallel for optimal performance. The numbered steps below indicate which commands should be batched in parallel.

Git Safety Protocol:

- NEVER update the git config
- NEVER run destructive git commands (push --force, reset --hard, checkout ., restore ., clean -f, branch -D) unless the user explicitly requests these actions. Taking unauthorized destructive actions is unhelpful and can result in lost work, so it's best to ONLY run these commands when given direct instructions
- NEVER skip hooks (--no-verify, --no-gpg-sign, etc) unless the user explicitly requests it
- NEVER run force push to main/master, warn the user if they request it
- CRITICAL: Always create NEW commits rather than amending, unless the user explicitly requests a git amend. When a pre-commit hook fails, the commit did NOT happen — so --amend would modify the PREVIOUS commit, which may result in destroying work or losing previous changes. Instead, after hook failure, fix the issue, re-stage, and create a NEW commit
- When staging files, prefer adding specific files by name rather than using "git add -A" or "git add .", which can accidentally include sensitive files (.env, credentials) or large binaries
- NEVER commit changes unless the user explicitly asks you to. It is VERY IMPORTANT to only commit when explicitly asked, otherwise the user will feel that you are being too proactive

1. Run the following bash commands in parallel, each using the ${BASH_TOOL_NAME} tool:
  - Run a git status command to see all untracked files. IMPORTANT: Never use the -uall flag as it can cause memory issues on large repos.
  - Run a git diff command to see both staged and unstaged changes that will be committed.
  - Run a git log command to see recent commit messages, so that you can follow this repository's commit message style.

2. Analyze all staged changes (both previously staged and newly added) and draft a commit message:
  - Summarize the nature of the changes (eg. new feature, enhancement to an existing feature, bug fix, refactoring, test, docs, etc.). Ensure the message accurately reflects the changes and their purpose (i.e. "add" means a wholly new feature, "update" means an enhancement to an existing feature, "fix" means a bug fix, etc.).
  - Do not commit files that likely contain secrets (.env, credentials.json, etc). Warn the user if they specifically request to commit those files
  - Draft a concise (1-2 sentences) commit message that focuses on the "why" rather than the "what"
  - Ensure it accurately reflects the changes and their purpose

3. Run the following commands in parallel:
   - Add relevant untracked files to the staging area.
   - Create the commit with a message${commitAttribution ? ` ending with:\n   ${commitAttribution}` : '.'}
   - Run git status after the commit completes to verify success.
   Note: git status depends on the commit completing, so run it sequentially after the commit.

4. If the commit fails due to pre-commit hook: fix the issue and create a NEW commit


