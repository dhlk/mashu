INTERNAL := \
	blend.py \
	demon-slayer-scrolls.blend \
	mashu \
	mashu-blend \
	mashu-blend-frame \
	mashu-stack

BINARYLINK := \
	mashu \
	mashu-blend \
	mashu-blend-frame \
	mashu-stack

all: mashu

install:
	install -Dm755 -t "$(DESTDIR)/$(PREFIX)/lib/mashu" $(INTERNAL)
	install -d $(DESTDIR)/$(PREFIX)/bin
	for tg in $(BINARYLINK); do \
		ln -s /$(PREFIX)/lib/mashu/$$tg $(DESTDIR)/$(PREFIX)/bin/$$tg; \
	done

mashu: catalog.go clip.go ffmpeg.go main.go stack.go struct.go
	gccgo -Wall -Werror $^ -lfsmap -o $@
