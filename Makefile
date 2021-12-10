clean:
	killall demory || true
	rm -rf /tmp/80*

bootstrap:
	DEMORY_NODE_ID=8080 \
	DEMORY_BOOTSTRAP=true \
	DEMORY_NODE_ADDRESS=localhost:8080 \
	DEMORY_PORT=8080 \
	DEMORY_DISCOVERY_STRATEGY=port go run ./cmd/.

cluster-9:
	number=8081 ; while [[ $$number -le 8090 ]] ; do \
		DEMORY_NODE_ID=$$number \
        DEMORY_BOOTSTRAP=false \
        DEMORY_NODE_ADDRESS=localhost:$$number \
        DEMORY_PORT=$$number \
        DEMORY_DISCOVERY_STRATEGY=port go run ./cmd/ & \
		((number = number + 1)) ; \
	done