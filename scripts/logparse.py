#!/usr/bin/env python3

import sys
import os

#parse log file
logfile = "./logfile.log"


# Find string in file and wite to csv file



def parse_log_file(log_file):
    with open(log_file, 'r') as f:
        for line in f:
            if '[FEP-CLIENT[AVALABLE]]' in line:
                fep="FEP AVAIALABLE"
            elif '[PROD-FLOWVOL-SERVER2[AVALABLE]]' in line:   
                flowvol="PROD FLOWVOL SERVER2 AVAIALABLE"
                with open('log.csv', 'a') as csv_file:
                    csv_file.write(fep, flowvol + '\n')


if __name__ == '__main__':

parse_log_file('./logfile.log')







