.PHONY: all tiles 
all: build

# Run brew install osmium-tool on macOS to download this tool
OSMIUM := $(shell which osmium)
# Build from sources at https://github.com/systemed/tilemaker
TILEMAKER_DIR := ../tilemaker
# Bounds can be selected f.e. by going to 
WGET := $(shell which wget)

tiles:
	mkdir -p work/tiles && \
	cd work && \
	wget http://download.geofabrik.de/europe/norway-latest.osm.pbf && \
	$(OSMIUM) extract \
		--bbox=9.78,63.17,10.97,63.56 \
		--set-bounds \
		--strategy=smart \
		norway-latest.osm.pbf \
		--output trondheim.osm.pbf && \
	$(TILEMAKER_DIR)/tilemaker \
		--input trondheim.osm.pbf \
		--output tiles \
		--process $(TILEMAKER_DIR)/resources/process-openmaptiles.lua \
		--config $(TILEMAKER_DIR)/resources/config-openmaptiles.json 
	rm -fR ../map/tiles && \
	mv tiles ../map 

	