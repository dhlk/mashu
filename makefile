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
	    json.go \
	    plangenerator.go \
	    project.go \
	    stack.go \
	    struct.go \
	    main.go

all: mashu

install: all
	install -Dm755 -t "$(DESTDIR)/$(PREFIX)/lib/mashu" $(INTERNAL)
	install -d $(DESTDIR)/$(PREFIX)/bin
	for tg in $(BINARYLINK); do \
		ln -s /$(PREFIX)/lib/mashu/$$tg $(DESTDIR)/$(PREFIX)/bin/$$tg; \
	done

mashu: $(MASHUSRC)
	gccgo -Wall -Werror $^ -lfsmap -o $@
