#!/usr/bin/python
import string
import random
from subprocess import call

new_password = ''.join(random.SystemRandom()
                       .choice(string.ascii_uppercase + string.digits)
                       for _ in range(16))

call("/bin/echo -e \"{}\\n{}\" | passwd myuser"
     .format(new_password, new_password), shell=True)

