JS_FILES = $(filter-out %min.js,$(wildcard js/*.js))
JS_MINIFIED= $(JS_FILES:.js=.min.js)

all: pauzer $(JS_MINIFIED)

pauzer: pauzer.go
	go build pauzer.go
	pkill pauzer
	./pauzer &

%.min.js: %.js
	uglifyjs $< -o $@

clean:
	rm -f $(JS_MINIFIED)
