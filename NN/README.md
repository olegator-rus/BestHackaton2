# Netcap Tensorflow Deep Neural Network
Author: Philipp Mieden (PLEXeT). For leaders of digital. August of 2020.

This repository contains a python implementation for using a Deep Neural Network with [Keras](https://keras.io) and [Tensorflow](https://www.tensorflow.org),
that operates on CSV data produced by the netcap framework.

Dataset: [CIC-IDS-2017](https://www.unb.ca/cic/datasets/ids-2017.html).
Each test is executed with a dedicated shell script.

First, the PCAP file is parsed with netcap,
in order to get audit records that will be labeled afterwards with the netlabel tool.
The labeled CSV data for the TCP audit record type is then used for training (75%) and evaluation (25%) of the classification accuracy provided by the deep neural network.


## Usage

    $ netcap-tf-dnn.py -h
    usage: netcap-tf-dnn.py [-h] -read READ [-drop DROP] [-sample [SAMPLE]]
                            [-dropna] [-string_dummy] [-string_index]
                            [-test_size TEST_SIZE] [-loss LOSS]
                            [-optimizer OPTIMIZER]

    NETCAP compatible implementation of Network Anomaly Detection with a Deep
    Neural Network and TensorFlow

    optional arguments:
    -h, --help            show this help message and exit
    -read READ            Labeled input CSV file to read from (required)
    -drop DROP            optionally drop specified columns, supply multiple
                            with comma
    -sample [SAMPLE]      optionally sample only a fraction of records
    -dropna               drop rows with missing values
    -string_dummy         encode strings as dummy variables
    -string_index         encode strings as indices (default)
    -test_size TEST_SIZE  specify size of the test data in percent (default:
                            0.25)
    -loss LOSS            set function (default: categorical_crossentropy)
    -optimizer OPTIMIZER  set optimizer (default: adam)

## License

Apache License 2.0