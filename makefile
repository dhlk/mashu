INTERNAL := \
	blend.py \
	demon-slayer-scrolls.blend \
	macros.vim

INTERNALBINARY := mashu

BINARYLINK := mashu

MASHUSRC := \
	    blend.go \
	    build.go \
	    catalog.go \
	    clip.go \
	    concat.go \
	    ffmpeg.go \
	    json.go \
	    m3u.go \
	    plangenerator.go \
	    project.go \
	    stack.go \
	    struct.go \
	    main.go

all: mashu

install: all
	install -Dm644 -t "$(DESTDIR)/$(PREFIX)/lib/mashu" $(INTERNAL)
	install -Dm755 -t "$(DESTDIR)/$(PREFIX)/lib/mashu" $(INTERNALBINARY)
	install -d $(DESTDIR)/$(PREFIX)/bin
	for tg in $(BINARYLINK); do \
		ln -s /$(PREFIX)/lib/mashu/$$tg $(DESTDIR)/$(PREFIX)/bin/$$tg; \
	done

mashu: $(MASHUSRC)
	gccgo -Wall -Werror $^ -lfsmap -o $@
