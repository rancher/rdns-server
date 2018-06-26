FROM python:2.7.14-stretch

ADD . /integration
WORKDIR /integration/tests

RUN pip install -r requirements.txt

ENTRYPOINT ["./run.sh"]