SHELL := /bin/bash

.PHONY: create-model-gemma3-weather
create-model-gemma3-weather:
	ollama create unplugged-gemma3-weather -f modelfiles/unplugged-gemma3-weather.modelfile

.PHONY: create-model-gemma3-isil
create-model-gemma3-isil:
	ollama create unplugged-gemma3-isil -f modelfiles/unplugged-gemma3-isil.modelfile


