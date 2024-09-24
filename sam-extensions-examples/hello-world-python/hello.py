import logging
def handler(event, context):
    logging.info("info log")
    logging.warning("warning log")
    logging.error("error log")
    print("Hello from python! to stderr")
    return "Hello from python! to stdout"