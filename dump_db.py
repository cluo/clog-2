#!/usr/bin/env python3

import struct
import sys

import plyvel

db = plyvel.DB(sys.argv[1], create_if_missing=False)
it = db.iterator()
for k, v in it:
	word = k.decode('utf-8')
	date, line_no = struct.unpack('LI', v)
	print(word, date, line_no)
