from setuptools import setup, find_packages
from codecs import open
from os import path

here = path.abspath(path.dirname(__file__))

# Get the long description from the README file
with open(path.join(here, 'README.md'), encoding='utf-8') as f:
    long_description = f.read()

setup(name='interline-healthcheck',
    version='0.0.0',
    description='Interline Healthcheck Tools',
    long_description=long_description,
    url='https://github.com/interline-io/interline-healthcheck',
    author='Ian Rees',
    author_email='ian@interline.io',
    packages=find_packages(exclude=['contrib', 'docs', 'tests']),
    install_requires=['requests'],
    zip_safe=True,
    package_data = {},
    classifiers=[
        'Intended Audience :: Developers',
    ],
    entry_points={
        'console_scripts': [
            'healthcheck=healthcheck:__main__',
        ],
    },
)
