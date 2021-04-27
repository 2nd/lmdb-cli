.PHONY: t
t:
	ruby test/populate.rb
	go test ./... -v
