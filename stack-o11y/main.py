from opentelemetry.sdk.resources import Resource

import logging
from opentelemetry._logs import set_logger_provider
from opentelemetry.exporter.otlp.proto.http._log_exporter import (
    OTLPLogExporter, 
)
from opentelemetry.sdk._logs import LoggerProvider, LoggingHandler
from opentelemetry.sdk._logs.export import (
    BatchLogRecordProcessor 
)
logger_provider = LoggerProvider(
    resource=Resource.create(
        {
        "lambda.name": "shoppingcart",
            "service.instance.id": "instance-12",
        }
    )
)

set_logger_provider(logger_provider)

exporter = OTLPLogExporter(endpoint="http://localhost:4318/v1/logs")
logger_provider.add_log_record_processor(BatchLogRecordProcessor(exporter))
handler = LoggingHandler(level=logging.INFO, logger_provider=logger_provider)

logging.getLogger().setLevel(logging.INFO)
logging.getLogger().addHandler(handler)
logger1 = logging.getLogger("myapp.area1")
print("LOGGER 1 LOG LEVEL IS: ", logger1.getEffectiveLevel())



def main():
    logger1.info("-#_#### un logardium :::::::::::")
    
    print('Started a root span')

    logger1.info("JOAKOOOOOOOOOO 1 INFO")
    print('Started a child span')

    logger1.info("JOAKOOOOOOOOOO 2 INFO")
    logger1.warning("JOAKOOOOOOOOOO child 2.1 WARN")
    logger1.error("JOAKOOOOOOOOOO child 2.1 ERROR")
        
    
    logger_provider.shutdown()


main()