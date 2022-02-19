FROM registry.access.redhat.com/ubi8-minimal

ENV BUILD_NAME svc
ENV SERVICE_NAME podops_api
ENV PODOPS_CONFIG_PATH /config/podops/config

EXPOSE 8080

# Copy extra files to the image
COPY ./run /root/usr/bin/

VOLUME /config
VOLUME /data

# copy the binary
COPY $BUILD_NAME /usr/bin/$SERVICE_NAME

# make it executable
RUN chmod +x /usr/bin/$SERVICE_NAME

CMD ["/root/usr/bin/run"]
