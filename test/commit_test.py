"""Test script."""

import os
import shutil
import subprocess
import time

TEST_PATH = './gtm_tests/'
GTM_PATH = '../build/go_build_github_com_kilpkonn_gtm_enhanced'
TEST_FILE_NAME = 'test.txt'


def setup():
    """Setup tests."""
    shutil.rmtree(TEST_PATH, ignore_errors=True)
    os.makedirs(TEST_PATH, exist_ok=True)
    shutil.copyfile(GTM_PATH, TEST_PATH + 'gtm')
    subprocess.call(['git', 'init'])
    with open(TEST_FILE_NAME, 'w') as f:
        f.write("0")
    subprocess.call(['git', 'add', '.'])
    subprocess.call(['git', 'commit', '.', '-m "test"'])


def cleanup():
    """Clean tests"""
    shutil.rmtree(TEST_PATH, ignore_errors=True)


def test_commit_10000_times_benchmark():
    """Test commit 10000 times"""
    for i in range(10000):
        with open(TEST_FILE_NAME, 'w') as f:
            f.write(str(i))
        subprocess.call(['git', 'commit', '.', '-m "test"'])


def test_commit_10000_times_gtm():
    """Test commit 10000 times"""
    for i in range(10000):
        with open(TEST_FILE_NAME, 'w') as f:
            f.write(str(i))
        subprocess.call(['gtm', 'record', TEST_FILE_NAME])
        subprocess.call(['git', 'commit', '.', '-m "test"'])


if __name__ == '__main__':
    os.makedirs(TEST_PATH, exist_ok=True)
    os.chdir(TEST_PATH)

    setup()
    start_benchmark = time.time()
    test_commit_10000_times_benchmark()
    end_benchmark = time.time()
    cleanup()

    setup()
    start_gtm = time.time()
    test_commit_10000_times_gtm()
    end_gtm = time.time()
    cleanup()

    print(f'Benchmark time: {round(end_benchmark - start_benchmark, 2)}s')
    print(f'Gtm time: {round(end_gtm - start_gtm, 2)}s')
