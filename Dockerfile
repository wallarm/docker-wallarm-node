FROM debian:jessie
MAINTAINER Wallarm Support Team <support@wallarm.com>

ENV DEBIAN_FRONTEND noninteractive

COPY conf/sources.list /etc/apt/sources.list.d/wallarm-node.list

RUN printf -- "-----BEGIN PGP PUBLIC KEY BLOCK-----\nVersion: GnuPG v1.4.12 (GNU/Linux)\n\nmQINBFL1Xl4BEADEFCVumPx2W4hQJG+4RRS0Zjw503a0YKH8tKp3OEWIMKiWwWia\nTcqxZghCZlm+MytwVmhX4pfEnkGyWdQTZOYosukTYqqYWnVEtqxTaep1k9JnUJ4r\nHsBUXIbnkL01rjLAkCxTTCMPzfQJNsqESnjllX6Ov/DtEm7EvilWdkkVK9TPF/tD\n0YwmIKz+nIR6Vylwy71f4hI6O2+91D5UJg/FardAner3rbIzYsRLgbAwx+5V8T9H\nSVWcjwAknpXDll6mvwionS5Aq+0hSuSjjABcZ2D7EW955ecb2Ql4fOEJPdmUQ//p\nHMHFgF85j4zwK5gfx9qLeTcmPxi1o14qmaKiZgfh7PxedScZP0VXN9B7Z9NFpo2Z\nQHqcBAK23msd+wnT4QMkC0CMLpw6AM0KmyNDVrVlTCytucg0zLBTBojEwSW6EzMS\nBVKip22qT4RelL3ykoHIHOoXbHPqSgdy+ba0A+gsfXmYHSOS/GjtNesgHTkUvyln\niZXyCGCy3Rqt+QzBWB2wXr2zQ13GJpnLprvtrVLx+GX0oHk6hF21Vm4iCFiw3eE8\ndI1y8wR2kDFjdwyduQPLXXPjt7aEdzCLoCaFRYju8k4jSGEaWa79jBvDer1fKhtH\nUmMz7HLmRRfKCSuJqMRJCU9spknt7g4e33OkKQprU+4Y5V7gtpMq1O/ECQARAQAB\ntEBXYWxsYXJtIERlYmlhbiBHTlUvTGludXggcGFja2FnZXMgcmVwb3NpdG9yeSA8\nYWRtaW5Ad2FsbGFybS5jb20+iQI+BBMBAgAoBQJS9V5eAhsDBQkJZgGABgsJCAcD\nAgYVCAIJCgsEFgIDAQIeAQIXgAAKCRAJY9VBcrhl/UWID/4pcIjqrzT3qM6SF9ow\naPmvB311hzghrzc7z3IEWn4gKZDF1Z9YMgWovTKZ+/0j3xzk66jnmcA2xnYOpKXL\nmrfiFC4QmQKbSoq4bKAiVx5q2nuUosxOUFODIvs5ORR4R75mW2I5f8aes/3+7vIx\nDEt9aYMEqqV7bc6/87lCqGMNQ6przIgRo2sXpt0drtK3TagjNwlDqihd7lx3/7VR\nt3G0Pi17rlfLGp4QjCeGoNmBRpLhM//USRR5J2sfFDFU09Zz/SV47/QMI2isUaeh\nHoAm3JHzC2RUDt8gpzjTop+ill/8Xbwr6Lh3WGsWRCsTlKAl2V9I5EiX+Hr8bTKT\nQG0pP1gBwnWb6zhThpBH+a303orsmu6GLOXjmTNjCNjlFdh8qW4+FgeZQuvC3m5o\nQC/7JBYrwuZFHTfsgoZtSKAoyy4IFZUCK3kbwuZLeKPW8GSrfXmo6Z+y5DFYQZcW\nTGpW+ogWrHJME8h0TmaID3fo3CVPX/MAhZfmeQy8Hb/NAw3/V7Dx5PWBguAbrNsw\njnf5w78RChPiCjRvA7wqK6eRCuw9Lg+C2Crqa5M67AEGd9tOwE315L27lM9gViH+\nPv4L0fJOtUsRyTftSx0nOLJF89XadRjmFVWZ5/1X8jaQ0vmNLTrSJw61vH6hS/oa\nAZ/zmKdciNja23uQ5HK+nhIquQ==\n=XjGi\n-----END PGP PUBLIC KEY BLOCK-----" | apt-key add - \
    && apt-get -y update && apt-get -y upgrade \
    && apt-get -y --no-install-recommends install \
        cron \
        logrotate \
        monitoring-plugins \
        supervisor \
        wallarm-node \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* \
    && chown -R wallarm:wallarm /var/lib/wallarm-tarantool \
    && sed -i -e 's|/var/log/wallarm/brute\.log|/var/log/wallarm/brute-detect\.log|' /etc/logrotate.d/wallarm-common \
    && sed -i -e 's|/usr/share/wallarm-common/syncnode|/usr/share/wallarm-common/syncnode -c /etc/wallarm-dist/node.yaml|' /etc/cron.d/wallarm-common \
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
