FROM python:2.7.14-stretch

ADD . /integration

WORKDIR /integration

RUN pip install -r requirements.txt

ENTRYPOINT ["./run.sh"]