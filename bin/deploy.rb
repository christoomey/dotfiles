#!/usr/bin/env ruby

class Deployer
  def initialize(environment, subcommands)
    @remote = environment
    @subcommands = subcommands
  end

  def run
    heroku_remote_targeted
    master_or_die
    working_directory_clean_or_die
    index_clean_or_die
    test_and_push
    warn_if_new_migrations
  end

  private

  def warn_if_new_migrations
    if system(%{git diff --name-status production/master..master | grep "^A" | cut -d$'\t' -f 2 | grep -q "db/migrate"})
      puts "\n\n\n" + "="*50 + "\n"
      puts "There are new migrations with this change"
    end
  end

  def heroku_remote_targeted
    unless heroku_remotes.include? @remote
      die "'#{@remote}' does not seem to be a heroku app"
    end
  end

  def heroku_remotes
    remotes = `git remote`.split.map(&:chomp)

    remotes.select do |remote|
      url = `git config --get remote.#{remote}.url`
      url =~ /heroku/
    end
  end

  def master_or_die
    if current_branch != 'master' && !off_master_switch_provided
      msg = "Expected branch for deployments is master. Current branch: [#{current_branch}]"
      msg += "\n   use -o/--off-master switch to deploy alternate branch"
      die msg
    elsif current_branch == 'master' && off_master_switch_provided
      die 'Current branch is master, but "--off-master" switch provided'
    end
  end

  def current_branch
    @branch || @branch = `git rev-parse --abbrev-ref HEAD`.chomp
  end

  def working_directory_clean_or_die
    git_clean
  end

  def index_clean_or_die
    git_clean(true)
  end

  def git_clean(index=false)
    if index
      clean = system("git diff --cached --exit-code > /dev/null 2>&1")
    else
      clean = system("git diff --exit-code > /dev/null 3>&1")
    end
    if !clean
      if index
        die "Index has unsaved changes. Aborting deploy."
      else
        die "Working directory has unsaved changes. Aborting deploy."
      end
    end
  end

  def off_master_switch_provided
    switches = ['-o', '--off-master']
    !(switches & @subcommands).empty?
  end

  def deploy
    if off_master_switch_provided
      deploy_cmd = "git push #{@remote} #{current_branch}:master"
    else
      deploy_cmd = "git push #{@remote} master"
    end
    system(deploy_cmd)
  end

  def test_and_push
    if system('bundle exec rake')
      deploy
    else
      die 'Rake exited with non zero status. Aborting commit.'
    end
  end

  def die(msg)
    puts msg
    exit 1
  end
end
