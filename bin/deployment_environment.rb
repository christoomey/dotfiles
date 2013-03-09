class DeploymentEnvironment
  def initialize(subcommands)
    @subcommands = subcommands
    @subcommand = @subcommands.first
  end

  def run
    case @subcommand
    when 'backup'
      system "heroku pgbackups:capture --expire --remote #{@environment}"
    when 'console'
      system "heroku run console --remote #{@environment}"
    when 'migrate'
      system %{
        heroku run rake db:migrate --remote #{@environment} &&
        heroku restart --remote #{@environment}
      }
    when 'tail'
      system "heroku logs --tail --remote #{@environment}"
    when 'url'
      system "heroku apps:info -r #{@environment} | grep -i web  | tr -s ' ' | cut -d ' ' -f 3"
    else
      system "heroku #{@subcommands.join(' ')} --remote #{@environment}"
    end
  end
end
