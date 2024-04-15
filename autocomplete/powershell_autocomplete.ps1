$fn = $($MyInvocation.MyCommand.Name)
$name = $fn -replace "(.*)\.ps1$", '$1'
Register-ArgumentCompleter -Native -CommandName $name -ScriptBlock {
    param($commandName, $wordToComplete, $cursorPosition)
    $other = "$wordToComplete --generate-shell-completion"
	Try {
    	Invoke-Expression $other | ForEach-Object {
        	[System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
    	}
	} Catch {
		Write-Error "Error generating completions: $_"
	}
 }