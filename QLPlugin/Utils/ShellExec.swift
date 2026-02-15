import Foundation

/// Result of executing an external process.
struct ExecResult {
	let stdout: String?
	let stderr: String?
	let exitCode: Int32
}

/// Error thrown when an external process exits with a non-zero status.
struct ExecError: Error {
	let execResult: ExecResult
}

/// Executes an external program and returns its output.
///
/// - Parameters:
///   - program: Absolute path to the executable.
///   - arguments: Command-line arguments to pass.
/// - Returns: An `ExecResult` containing stdout, stderr, and the exit code.
/// - Throws: `ExecError` if the process exits with a non-zero status.
func exec(program: String, arguments: [String] = []) throws -> ExecResult {
	let process = Process()
	process.executableURL = URL(fileURLWithPath: program)
	process.arguments = arguments

	let stdoutPipe = Pipe()
	let stderrPipe = Pipe()
	process.standardOutput = stdoutPipe
	process.standardError = stderrPipe

	try process.run()
	process.waitUntilExit()

	let stdoutData = stdoutPipe.fileHandleForReading.readDataToEndOfFile()
	let stderrData = stderrPipe.fileHandleForReading.readDataToEndOfFile()

	let result = ExecResult(
		stdout: String(data: stdoutData, encoding: .utf8),
		stderr: String(data: stderrData, encoding: .utf8),
		exitCode: process.terminationStatus
	)

	if result.exitCode != 0 {
		throw ExecError(execResult: result)
	}

	return result
}
