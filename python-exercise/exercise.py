#!/usr/local/opt/python@3.8/bin/python3

import sys
import datetime

# list of numbers comma separated

list_of_numbers = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
print(list_of_numbers)

print(tuple(list_of_numbers))

# convert date into days
date1 = datetime.datetime(2014, 7, 2)
date2 = datetime.datetime(2014, 7, 11)
print(date2 - date1)


# print the current time
