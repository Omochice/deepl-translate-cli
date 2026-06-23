$fn = $($MyInvocation.MyCommand.Name)
$name = $fn -replace "(.*)\.ps1$", '$1'
Register-ArgumentCompleter -Native -CommandName $name -ScriptBlock {
    param($commandName, $wordToComplete, $cursorPosition)
    $other = "$wordToComplete --generate-shell-completion"
	# Use array parameters with call operator (&) to pass arguments securely:
	Try {
	  & $name --generate-shell-completion $wordToComplete | ForEach-Object {
		  [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
	  }
	} Catch {
	  Write-Error "Error generating completions: $_"
	}
 }