INTERNAL := \
	blend.py \
	demon-slayer-scrolls.blend \
	mashu-args \
	mashu-blend \
	mashu-blend-frame \
	mashu-clip \
	mashu-stack

BINARYLINK := \
	mashu-blend \
	mashu-blend-frame \
	mashu-clip \
	mashu-stack

install:
	install -Dm755 -t "$(DESTDIR)/$(PREFIX)/lib/mashu" $(INTERNAL)
	install -d $(DESTDIR)/$(PREFIX)/bin
	for tg in $(BINARYLINK); do \
		ln -s /$(PREFIX)/lib/mashu/$$tg $(DESTDIR)/$(PREFIX)/bin/$$tg; \
	done
