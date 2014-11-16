FROM golang:1.3.3-onbuild

EXPOSE 4001

# install confd and configure
WORKDIR /usr/local/bin
RUN curl -L https://github.com/kelseyhightower/confd/releases/download/v0.5.0/confd-0.5.0-linux-amd64 -o confd
RUN chmod +x confd
RUN ["/bin/bash", "-c", "mkdir -p /etc/confd/{conf.d,templates}"]
ADD ./confd/km.tmpl /etc/confd/templates/km.tmpl
ADD ./confd/km.toml /etc/confd/conf.d/km.toml
ADD ./confd/confd-watch /usr/local/bin/confd-watch
RUN chmod +x /usr/local/bin/confd-watch

CMD /usr/local/bin/confd-watch
