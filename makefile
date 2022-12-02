INTERNAL := \
	blend.py \
	demon-slayer-scrolls.blend \
	mashu

BINARYLINK := mashu

MASHUSRC := \
	    blend.go \
	    catalog.go \
	    clip.go \
	    concat.go \
	    ffmpeg.go \
	    stack.go \
	    struct.go \
	    stream.go \
	    main.go

all: mashu

install:
	install -Dm755 -t "$(DESTDIR)/$(PREFIX)/lib/mashu" $(INTERNAL)
	install -d $(DESTDIR)/$(PREFIX)/bin
	for tg in $(BINARYLINK); do \
		ln -s /$(PREFIX)/lib/mashu/$$tg $(DESTDIR)/$(PREFIX)/bin/$$tg; \
	done

mashu: $(MASHUSRC)
	gccgo -Wall -Werror $^ -lfsmap -o $@
