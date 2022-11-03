FROM centos:centos7

RUN mkdir /app
WORKDIR /app
RUN yum install -y python pip  python3-pip
RUN python -m pip install virtualenv
RUN python -m ensurepip --default-pip
RUN python -m  virtualenv venv  .venv && .venv/bin/pip install --no-cache-dir -U pip setuptools
COPY        src/ /app/
COPY *whl /tmp/
RUN yum install -y  gcc && \
	#.venv/bin/pip install --no-cache-dir -r requirements.txt && \
	.venv/bin/pip install --no-cache-dir -r requirements.txt --no-index --find-links /tmp && \
    rm requirements.txt && \
	find /app/.venv \( -type d -a -name test -o -name tests \) -o \( -type f -a -name '*.pyc' -o -name '*.pyo' \) -exec rm -rf '{}' \+

ENV PYTHONUNBUFFERED=1
WORKDIR /app
ENV PATH="/app/.venv/bin:$PATH"
USER        65534:65534
CMD         [ "python", "-u", "/app/sidecar.py" ]
