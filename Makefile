

NAME := countnsort


install:
	go build ./main.go
	ln -s "${PWD}/main" "${HOME}/.local/bin/$(NAME)"

uninstall:
	rm "${HOME}/.local/bin/$(NAME)"
