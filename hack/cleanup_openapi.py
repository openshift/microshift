#!/usr/bin/env python3

import json
import re
import sys


url_pat = re.compile('#\S+"$')

def fix_values(data, location=''):
    if isinstance(data, str):
        # Remove trailing quotes from strings like
        #
        #   Kind of the referent; More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds"
        #
        # so the conversion from asciidoc to DocBook does not result
        # in invalid XML.
        if url_pat.search(data):
            print(f'fixing {location}')
            data = data.rstrip('"')
        return data

    if not isinstance(data, dict):
        return data

    for key, value in list(data.items()):
        data[key] = fix_values(value, location + '.' + key)

    return data

def main():
    if len(sys.argv) != 3:
        raise SystemExit('script requires 2 arguments, input_file and output_file')
    input_filename = sys.argv[1]
    output_filename = sys.argv[2]

    with open(input_filename, 'r') as inf:
        data = json.load(inf)

    data = fix_values(data)

    with open(output_filename, 'w') as outf:
        json.dump(data, outf)


if __name__ == '__main__':
    main()
