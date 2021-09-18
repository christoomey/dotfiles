alias be="bundle exec "

rspec_failures() {
  grep 'failed' spec/examples.txt | \
    awk '{print $1}' | \
    sed 's/\[.*$//' | \
    sed 's/\.\///' | \
    sort | \
    uniq
}

export OBJC_DISABLE_INITIALIZE_FORK_SAFETY=YES
