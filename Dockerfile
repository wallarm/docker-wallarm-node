FROM debian:stretch
MAINTAINER Wallarm Support Team <support@wallarm.com>

ENV DEBIAN_FRONTEND noninteractive

COPY conf/sources.list /etc/apt/sources.list.d/wallarm-node.list

RUN printf -- "mQINBFL1Xl4BEADEFCVumPx2W4hQJG+4RRS0Zjw503a0YKH8tKp3OEWIMKiWwWiaTcqxZghCZlm+MytwVmhX4pfEnkGyWdQTZOYosukTYqqYWnVEtqxTaep1k9JnUJ4rHsBUXIbnkL01rjLAkCxTTCMPzfQJNsqESnjllX6Ov/DtEm7EvilWdkkVK9TPF/tD0YwmIKz+nIR6Vylwy71f4hI6O2+91D5UJg/FardAner3rbIzYsRLgbAwx+5V8T9HSVWcjwAknpXDll6mvwionS5Aq+0hSuSjjABcZ2D7EW955ecb2Ql4fOEJPdmUQ//pHMHFgF85j4zwK5gfx9qLeTcmPxi1o14qmaKiZgfh7PxedScZP0VXN9B7Z9NFpo2ZQHqcBAK23msd+wnT4QMkC0CMLpw6AM0KmyNDVrVlTCytucg0zLBTBojEwSW6EzMSBVKip22qT4RelL3ykoHIHOoXbHPqSgdy+ba0A+gsfXmYHSOS/GjtNesgHTkUvylniZXyCGCy3Rqt+QzBWB2wXr2zQ13GJpnLprvtrVLx+GX0oHk6hF21Vm4iCFiw3eE8dI1y8wR2kDFjdwyduQPLXXPjt7aEdzCLoCaFRYju8k4jSGEaWa79jBvDer1fKhtHUmMz7HLmRRfKCSuJqMRJCU9spknt7g4e33OkKQprU+4Y5V7gtpMq1O/ECQARAQABtEBXYWxsYXJtIERlYmlhbiBHTlUvTGludXggcGFja2FnZXMgcmVwb3NpdG9yeSA8YWRtaW5Ad2FsbGFybS5jb20+iQI+BBMBAgAoBQJS9V5eAhsDBQkJZgGABgsJCAcDAgYVCAIJCgsEFgIDAQIeAQIXgAAKCRAJY9VBcrhl/UWID/4pcIjqrzT3qM6SF9owaPmvB311hzghrzc7z3IEWn4gKZDF1Z9YMgWovTKZ+/0j3xzk66jnmcA2xnYOpKXLmrfiFC4QmQKbSoq4bKAiVx5q2nuUosxOUFODIvs5ORR4R75mW2I5f8aes/3+7vIxDEt9aYMEqqV7bc6/87lCqGMNQ6przIgRo2sXpt0drtK3TagjNwlDqihd7lx3/7VRt3G0Pi17rlfLGp4QjCeGoNmBRpLhM//USRR5J2sfFDFU09Zz/SV47/QMI2isUaehHoAm3JHzC2RUDt8gpzjTop+ill/8Xbwr6Lh3WGsWRCsTlKAl2V9I5EiX+Hr8bTKTQG0pP1gBwnWb6zhThpBH+a303orsmu6GLOXjmTNjCNjlFdh8qW4+FgeZQuvC3m5oQC/7JBYrwuZFHTfsgoZtSKAoyy4IFZUCK3kbwuZLeKPW8GSrfXmo6Z+y5DFYQZcWTGpW+ogWrHJME8h0TmaID3fo3CVPX/MAhZfmeQy8Hb/NAw3/V7Dx5PWBguAbrNswjnf5w78RChPiCjRvA7wqK6eRCuw9Lg+C2Crqa5M67AEGd9tOwE315L27lM9gViH+Pv4L0fJOtUsRyTftSx0nOLJF89XadRjmFVWZ5/1X8jaQ0vmNLTrSJw61vH6hS/oaAZ/zmKdciNja23uQ5HK+nhIquQ==" | base64 -d > /usr/share/keyrings/wallarm.gpg \
    && apt-get -y update && apt-get -y upgrade \
    && apt-get -y --no-install-recommends install \
        cron \
        logrotate \
        monitoring-plugins \
        supervisor \
        nginx-wallarm \
        wallarm-node \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* \
    && chown -R wallarm:wallarm /var/lib/wallarm-tarantool \
    && sed -i -e 's|/var/log/wallarm/brute\.log|/var/log/wallarm/brute-detect\.log|' /etc/logrotate.d/wallarm-common \
    && sed -i -e 's|/usr/share/wallarm-common/syncnode|/usr/share/wallarm-common/syncnode -c /etc/wallarm-dist/node.yaml|' /etc/cron.d/wallarm-node-nginx \
    && rm -rf /etc/wallarm/triggers.d/ \
    && mkdir -p /etc/wallarm-dist/triggers.d \
    && sed -i -e '\@<Plugin syslog>@,\@</Plugin>@ s/^/#/' /etc/collectd/collectd.conf

COPY conf/node.yaml /etc/wallarm-dist/
COPY scripts/trigger /etc/wallarm-dist/triggers.d/nginx
COPY scripts/init /usr/local/bin/
COPY conf/supervisord.conf /etc/supervisor/
COPY conf/logrotate.conf /etc/

EXPOSE 80 443

CMD ["/usr/local/bin/init"]
