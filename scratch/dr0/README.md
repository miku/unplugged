# minimal deep research (dr0)

> tools to access various catalogs and databases

* search result, compressed; guess what whould be relevant
* "create a research plan", markdown, todo list
* take first open item, do query or ask for clarification; run query, read, update todo list, add summary

```
$ cat full.json | ./minimize.sh
{"title":"Recovery of deep research vehicle Alvin","author":["United States. Naval Ship Systems Command."],"year":null,"url":["https://www.biodiversitylibrary.org/item/86720"]}
{"title":"Recovery of deep research vehicle Alvin","author":["United States. Naval Ship Systems Command."],"year":"1969","url":["https://www.biodiversitylibrary.org/bibliography/39120"]}
{"title":"\"Deep Research\": A Research Paradigm Shift","author":["Shiri, Ali"],"year":"2025","url":["https://doi.org/10.2139/ssrn.5224935"]}
{"title":"A Deep Research in Classic Classification Network","author":["He, Hanyue"],"year":"2020","url":["https://doi.org/10.1088/1757-899x/740/1/012152"]}
{"title":"Deeper Research Possibilities on Nez Perce Hostage Suggested","author":["Brown, Mark H."],"year":"1969","url":["https://www.jstor.org/stable/4517341"]}
{"title":"AI’s deep research revolution: Transforming biomedical literature analysis","author":["Wang, Ying-Mei","Chen, Tzeng-Ji"],"year":"2025","url":["https://doi.org/10.1097/jcma.0000000000001231"]}
{"title":"Deep Research on Self-Inductance Issues Based on Experimental Phenomena","author":["谭, 文"],"year":"2024","url":["http://dx.doi.org/10.12677/mp.2024.144015"]}
{"title":"The popular music band, BTS in Iran: A deep research","author":["E. Shaghasemi"],"year":"2024","url":["https://doi.org/10.22035/jicr.2024.3289.3570"]}
{"title":"Why so deep research on Yb3+-doped optical inorganic materials?","author":["Boulon, Georges"],"year":"2008","url":["https://doi.org/10.1016/j.jallcom.2007.04.148"]}
{"title":"OpenAI’s ‘deep research’ tool: is it useful for scientists?","author":["Jones, Nicola"],"year":"2025","url":["https://doi.org/10.1038/d41586-025-00377-9"]}
```

Ask for next possily relevant question.

```
$ ollama run qwen3:14b --verbose
>>> """
... I am researching "deep research" tools for scientific research, and found following entries in a catalog.
... {"title":"Recovery of deep research vehicle Alvin","author":["United States. Naval Ship Systems Command."],"year":null,"url":["https://www.biodiversitylibrary.org/item/86720"]}
...
... {"title":"Recovery of deep research vehicle Alvin","author":["United States. Naval Ship Systems Command."],"year":"1969","url":["https://www.biodiversitylibrary.org/bibliograph
... y/39120"]}
... {"title":"\"Deep Research\": A Research Paradigm Shift","author":["Shiri, Ali"],"year":"2025","url":["https://doi.org/10.2139/ssrn.5224935"]}
... {"title":"A Deep Research in Classic Classification Network","author":["He, Hanyue"],"year":"2020","url":["https://doi.org/10.1088/1757-899x/740/1/012152"]}
... {"title":"Deeper Research Possibilities on Nez Perce Hostage Suggested","author":["Brown, Mark H."],"year":"1969","url":["https://www.jstor.org/stable/4517341"]}
... {"title":"AI’s deep research revolution: Transforming biomedical literature analysis","author":["Wang, Ying-Mei","Chen, Tzeng-Ji"],"year":"2025","url":["https://doi.org/10.1097
... /jcma.0000000000001231"]}
... {"title":"Deep Research on Self-Inductance Issues Based on Experimental Phenomena","author":["谭, 文"],"year":"2024","url":["http://dx.doi.org/10.12677/mp.2024.144015"]}
... {"title":"The popular music band, BTS in Iran: A deep research","author":["E. Shaghasemi"],"year":"2024","url":["https://doi.org/10.22035/jicr.2024.3289.3570"]}
... {"title":"Why so deep research on Yb3+-doped optical inorganic materials?","author":["Boulon, Georges"],"year":"2008","url":["https://doi.org/10.1016/j.jallcom.2007.04.148"]}
... {"title":"OpenAI’s ‘deep research’ tool: is it useful for scientists?","author":["Jones, Nicola"],"year":"2025","url":["https://doi.org/10.1038/d41586-025-00377-9"]}
... which look suitable for the question?
... only output the titles of the relevant entries, nothing more
... """

Thinking...
Okay, let's see. The user is looking for entries related to "deep research" tools for scientific research. They provided a list of entries, and I need to determine which ones
are relevant.

First, I'll go through each entry one by one. The first two entries are about the recovery of the Alvin vehicle, which seems to be a submersible. The titles mention "Recovery
of deep research vehicle Alvin," but the term "deep research" here might refer to the vehicle's purpose rather than a methodology or tool. The years are 1969 and null, and
the authors are from the Naval Ship Systems Command. These might not be about tools for research but rather about the vehicle itself. Probably not relevant.

Next, the third entry: "Deep Research: A Research Paradigm Shift" by Shiri, Ali, from 2025. The title directly mentions "deep research" and a paradigm shift, which could be
discussing methodologies or tools. This seems relevant.

Fourth entry: "A Deep Research in Classic Classification Network" by He, Hanyue, 2020. The title mentions "deep research" in the context of a classification network, which is
a machine learning topic. This might be about applying deep learning techniques, so possibly relevant to tools in scientific research.

Fifth entry: "Deeper Research Possibilities on Nez Perce Hostage Suggested" by Brown, Mark H., 1969. The title here seems to be about a specific historical event, not a tool
or methodology. Probably not relevant.

Sixth entry: "AI’s deep research revolution: Transforming biomedical literature analysis" by Wang and Chen, 2025. This title clearly connects AI with deep research in
biomedical literature analysis, which is a tool or method for scientific research. Definitely relevant.

Seventh entry: "Deep Research on Self-Inductance Issues Based on Experimental Phenomena" by 谭, 文, 2024. This seems to be a study on self-inductance, which is an engineering
topic. The term "deep research" here might refer to in-depth study rather than a tool. Maybe less relevant, but could be about methodologies.

Eighth entry: "The popular music band, BTS in Iran: A deep research" by E. Shaghasemi, 2024. This is about a music band, so likely a cultural study. The term "deep research"
here is probably not about tools but a thorough study. Not relevant.

Ninth entry: "Why so deep research on Yb3+-doped optical inorganic materials?" by Boulon, Georges, 2008. This is a specific materials science study. The title suggests a
focus on the material rather than a tool. Not relevant.

Tenth entry: "OpenAI’s ‘deep research’ tool: is it useful for scientists?" by Jones, Nicola, 2025. The title explicitly mentions a tool called "deep research" by OpenAI,
which is directly related to the user's query about tools for scientific research. This is highly relevant.

So, the relevant entries are the ones that discuss methodologies, tools, or paradigms related to deep research in scientific contexts. The third, fourth, sixth, seventh, and
tenth entries seem to fit. Wait, the seventh entry is about self-inductance based on experiments. The term "deep research" here might be a translation issue or a specific
methodology. It's a bit ambiguous, but since the user is looking for tools, maybe it's not about a tool but a study. The fourth entry is about a classification network, which
is a tool. The third entry is a paradigm shift, which could be a methodology. The sixth is about AI transforming literature analysis, which is a tool. The tenth is about a
specific tool by OpenAI. The seventh might be a stretch, but I'll include it unless there's a clear reason to exclude. Let me check again. The seventh title says "Deep
Research on Self-Inductance Issues Based on Experimental Phenomena." It's a study on a specific topic, not a tool. So maybe not relevant. So the relevant ones are 3, 4, 6,
10. Wait, entry 4: "A Deep Research in Classic Classification Network" – this might be about using deep learning in classification, which is a tool. So yes, relevant. Entry 3
is about a paradigm shift, which could be a new approach or tool. Entry 6 is about AI tools. Entry 10 is about a specific tool. So those four. Let me confirm again. The user
is looking for "deep research" tools, so entries that are about tools or methodologies that facilitate deep research. Entry 3 might be a theoretical paper on the paradigm
shift, which could be relevant. Entry 4 is about applying deep research in a classification network, which is a tool. Entry 6 is about AI's role in transforming literature
analysis, which is a tool. Entry 10 is about a tool by OpenAI. Entry 7 is about a study on self-inductance, not a tool. So the relevant titles are: "Deep Research: A Research
Paradigm Shift", "A Deep Research in Classic Classification Network", "AI’s deep research revolution: Transforming biomedical literature analysis", and "OpenAI’s ‘deep
research’ tool: is it useful for scientists?".
...done thinking.

"\"Deep Research\": A Research Paradigm Shift"
"A Deep Research in Classic Classification Network"
"AI’s deep research revolution: Transforming biomedical literature analysis"
"OpenAI’s ‘deep research’ tool: is it useful for scientists?"

total duration:       45.301374477s
load duration:        87.032343ms
prompt eval count:    647 token(s)
prompt eval duration: 65.532935ms
prompt eval rate:     9872.90 tokens/s
eval count:           1183 token(s)
eval duration:        44.826752918s
eval rate:            26.39 tokens/s
```

> Could that be an EVAL?

Given a research question and a list of results, determine the relevant
entries.
