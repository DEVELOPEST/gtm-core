"""Test script."""

import os
import stat
import shutil
import subprocess
import time

from matplotlib import pyplot as plt

TEST_PATH = './gtm_tests/'
GTM_PATH = '../build/go_build_github_com_kilpkonn_gtm_enhanced'
TEST_FILE_NAME = 'test.txt'


def setup():
    """Setup tests."""
    shutil.copyfile(GTM_PATH, 'gtm')
    st = os.stat('gtm')
    os.chmod('gtm', st.st_mode | stat.S_IEXEC)
    subprocess.call(['git', 'init'])
    with open(TEST_FILE_NAME, 'w') as f:
        f.write("0")
    subprocess.call(['git', 'add', '.'])
    subprocess.call(['git', 'commit', '.', '-m "test"'])


def cleanup():
    """Clean tests"""
    os.chdir('..')
    shutil.rmtree(TEST_PATH, ignore_errors=True)
    os.makedirs(TEST_PATH, exist_ok=True)
    os.chdir(TEST_PATH)


def get_size(start_path: str = '.'):
    total_size = 0
    for dir_path, dir_names, filenames in os.walk(start_path):
        for f in filenames:
            fp = os.path.join(dir_path, f)
            # skip if it is symbolic link
            if not os.path.islink(fp):
                total_size += os.path.getsize(fp)

    return total_size


def test_commit_benchmark(n: int):
    """Test commit 10000 times"""
    for i in range(n):
        with open(TEST_FILE_NAME, 'w') as f:
            f.write(str(i))
        subprocess.call(['git', 'commit', '.', '-m "test"'], )


def test_commit_gtm(n: int):
    """Test commit 10000 times"""
    for i in range(n):
        with open(TEST_FILE_NAME, 'w') as f:
            f.write(str(i))
        subprocess.call(['./gtm', 'record', TEST_FILE_NAME])
        subprocess.call(['git', 'commit', '.', '-m "test"'])


def test_commit_increasing_size_benchmark(n: int, increase_multiplier: float, offset: int = 0):
    """Test commit 10000 times with increasing file size."""
    for i in range(n):
        with open(TEST_FILE_NAME, 'a') as f:
            new_text = ''.join([f'{n + offset} - {x}\n' for x in range(round((offset + i) * increase_multiplier))])
            f.write(new_text)
        subprocess.call(['git', 'commit', '.', '-m "test"'])


def test_commit_increasing_size_gtm(n: int, increase_multiplier: float, offset: int = 0):
    """Test commit 10000 times with increasing file size."""
    for i in range(n):
        with open(TEST_FILE_NAME, 'a') as f:
            new_text = ''.join([f'{offset + n} - {x}\n' for x in range(round((offset + i) * increase_multiplier))])
            f.write(new_text)

        for _ in range(round((offset + i) * increase_multiplier)):
            subprocess.call(['./gtm', 'record', TEST_FILE_NAME])
        subprocess.call(['git', 'commit', '.', '-m "test"'])


def test_commit_record_gtm(n: int):
    """Test commit 10000 times"""
    for i in range(n):
        subprocess.call(['./gtm', 'record', TEST_FILE_NAME])


if __name__ == '__main__':
    os.makedirs(TEST_PATH, exist_ok=True)
    os.chdir(TEST_PATH)

    results = ['Results:']

    commit_times_benchmark = []
    commit_times_gtm = []
    git_size_benchmark = []
    git_size_gtm = []
    n = 5000
    x_tick_len = 10

    setup()
    for _ in range(n // x_tick_len):
        start_benchmark = time.time()
        test_commit_benchmark(x_tick_len)
        end_benchmark = time.time()
        commit_times_benchmark.append(round(end_benchmark - start_benchmark, 3))
        git_size_benchmark.append(get_size('.git'))
    size_benchmark = get_size('.git')
    cleanup()

    setup()
    for _ in range(n // x_tick_len):
        start_gtm = time.time()
        test_commit_gtm(x_tick_len)
        end_gtm = time.time()
        commit_times_gtm.append(round(end_gtm - start_gtm, 3))
        git_size_gtm.append(get_size('.git'))
    size_gtm = get_size('.git')
    cleanup()

    results.append('-' * 50)
    results.append(f'2000 commits Benchmark time: {round(sum(commit_times_benchmark), 2)}s')
    results.append(f'2000 commits Gtm time: {round(sum(commit_times_gtm), 2)}s')
    results.append(f'Benchmark .git folder size: {round(size_benchmark / 1024, 2)} kB')
    results.append(f'Gtm .git folder size: {round(size_gtm / 1024, 2)} kB')

    plt.plot([x for x in range(round(n / x_tick_len))], commit_times_benchmark, label='Benchmark')
    plt.plot([x for x in range(round(n / x_tick_len))], commit_times_gtm, label='Gtm')
    plt.legend()
    plt.xlabel('Run number')
    plt.ylabel(f'{x_tick_len} commit time')
    plt.savefig('../commit_times.png')
    plt.clf()

    plt.plot([x for x in range(round(n / x_tick_len))], git_size_benchmark, label='Benchmark')
    plt.plot([x for x in range(round(n / x_tick_len))], git_size_gtm, label='Gtm')
    plt.legend()
    plt.xlabel('Run number')
    plt.ylabel(f'{x_tick_len} .git folder size')
    plt.savefig('../git_size.png')
    plt.clf()

    commit_times_benchmark = []
    commit_times_gtm = []
    git_size_benchmark = []
    git_size_gtm = []
    n = 200
    x_tick_len = 5

    setup()
    for a in range(n // x_tick_len):
        start_benchmark = time.time()
        test_commit_increasing_size_benchmark(x_tick_len, 1.5, a * x_tick_len)
        end_benchmark = time.time()
        commit_times_benchmark.append(round(end_benchmark - start_benchmark, 3))
        git_size_benchmark.append(get_size('.git'))
    size_benchmark = get_size('.git')
    cleanup()

    setup()
    for a in range(n // x_tick_len):
        start_gtm = time.time()
        test_commit_increasing_size_gtm(x_tick_len, 1.5, a * x_tick_len)
        end_gtm = time.time()
        commit_times_gtm.append(round(end_gtm - start_gtm, 3))
        git_size_gtm.append(get_size('.git'))
    size_gtm = get_size('.git')
    cleanup()

    results.append('-' * 50)
    results.append(f'100 commits inc size Benchmark time: {round(sum(commit_times_benchmark), 2)}s')
    results.append(f'100 commits inc size (7500 record events) Gtm time: {round(sum(commit_times_gtm), 2)}s')
    results.append(f'Benchmark .git folder size: {round(size_benchmark / 1024, 2)} kB')
    results.append(f'Gtm .git folder size: {round(size_gtm / 1024, 2)} kB')

    record_times = []
    setup()
    for a in range(n // x_tick_len):
        start_gtm = time.time()
        test_commit_record_gtm(round(a * 1.5 * x_tick_len))
        end_gtm = time.time()
        record_times.append(round(end_gtm - start_gtm, 3))
    cleanup()

    commit_times_benchmark = [x - y for x, y in zip(commit_times_gtm, record_times)]
    plt.plot([x for x in range(round(n / x_tick_len))], commit_times_benchmark, label='Benchmark')
    plt.plot([x for x in range(round(n / x_tick_len))], commit_times_gtm, label='Gtm')
    plt.legend()
    plt.xlabel('Run number')
    plt.ylabel(f'{x_tick_len} commit time')
    plt.savefig('../commit_times_inc.png')
    plt.clf()

    plt.plot([x for x in range(round(n / x_tick_len))], git_size_benchmark, label='Benchmark')
    plt.plot([x for x in range(round(n / x_tick_len))], git_size_gtm, label='Gtm')
    plt.legend()
    plt.xlabel('Run number')
    plt.ylabel(f'{x_tick_len} .git folder size')
    plt.savefig('../git_size_inc.png')
    plt.clf()

    results.append(f'7500 record event time: {round(sum(record_times), 2)}s')

    for line in results:
        print(line)
