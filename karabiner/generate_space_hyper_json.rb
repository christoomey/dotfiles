require "json"

keys = ('a'..'z').to_a + ('0'..'9').to_a + ['open_bracket', 'close_bracket', 'comma', 'period', 'backslash']


rules = keys.map do |key|
  {
    'type' => 'basic',
    'conditions' => [
      {
        'type' => 'variable_if',
        'name' => 'hyper_key_mode',
        'value' => 1
      }
    ],
    'from' => {
      'key_code' => key,
      'modifiers' => {
        'optional' => ['any']
      }
    },
    'to' => [
      {
        'key_code' => key,
        'modifiers' => ['left_command', 'left_option', 'left_control']
      }
    ]
  }
end

config = {
  'title' => 'Space plus Key based Hyper Key Configuration',
  'rules' => [
    {
      'description' => 'Map space + key to left_command+left_option+left_control + key',
      'manipulators' => [
        {
          'type' => 'basic',
          'from' => {
            'key_code' => 'spacebar',
            'modifiers' => {
              'optional' => ['any']
            }
          },
          'to' => [
            {
              'set_variable' => {
                'name' => 'hyper_key_mode',
                'value' => 1
              }
            }
          ],
          'to_after_key_up' => [
            {
              'set_variable' => {
                'name' => 'hyper_key_mode',
                'value' => 0
              }
            }
          ],
          'to_if_alone' => [
            {
              'key_code' => 'spacebar'
            }
          ]
        }
      ] + rules
    }
  ]
}

File.write('space-hyper.json', JSON.pretty_generate(config))
