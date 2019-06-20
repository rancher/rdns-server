TARGETS := $(shell ls scripts)

.dapper:
	@echo Downloading dapper
	@curl -sL https://releases.rancher.com/dapper/latest/dapper-`uname -s`-`uname -m` > .dapper.tmp
	@@chmod +x .dapper.tmp
	@./.dapper.tmp -v
	@mv .dapper.tmp .dapper

$(TARGETS): .dapper
	./.dapper $@

mod: .dapper
	./.dapper -m bind go mod vendor

shell-bind: .dapper
	./.dapper -m bind -s

deps: trash

.DEFAULT_GOAL := ci

.PHONY: $(TARGETS)
