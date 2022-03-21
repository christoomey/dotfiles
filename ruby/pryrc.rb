def pbcopy(input)
  input.to_s.tap do |str|
    IO.popen('pbcopy', 'w') { |f| f << str }
  end
end
