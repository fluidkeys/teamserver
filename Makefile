MAIN_GO_FILES=main.go \
		 teamshandler.go \
		 summaryhandler.go \
		 requesthandler.go \

.PHONY: run
run: $(MAIN_GO_FILES)
	go run $(MAIN_GO_FILES)
