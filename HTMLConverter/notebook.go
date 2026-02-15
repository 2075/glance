// Jupyter Notebook to HTML converter.
// Replaces the archived github.com/samuelmeuli/nbtohtml dependency.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	htmlFormatter "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/microcosm-cc/bluemonday"
)

// --- Jupyter Notebook JSON types (nbformat v4) ---
// https://nbformat.readthedocs.io

type nbOutputData struct {
	TextHTML       []string `json:"text/html,omitempty"`
	ApplicationPDF *string  `json:"application/pdf,omitempty"`
	TextLaTeX      *string  `json:"text/latex,omitempty"`
	ImageSVGXML    []string `json:"image/svg+xml,omitempty"`
	ImagePNG       *string  `json:"image/png,omitempty"`
	ImageJPEG      *string  `json:"image/jpeg,omitempty"`
	TextMarkdown   []string `json:"text/markdown,omitempty"`
	TextPlain      []string `json:"text/plain,omitempty"`
}

type nbOutput struct {
	OutputType     string       `json:"output_type"`
	ExecutionCount *int         `json:"execution_count,omitempty"`
	Text           []string     `json:"text,omitempty"`
	Data           nbOutputData `json:"data,omitempty"`
	Traceback      []string     `json:"traceback,omitempty"`
}

type nbCell struct {
	CellType       string     `json:"cell_type"`
	ExecutionCount *int       `json:"execution_count,omitempty"`
	Source         []string   `json:"source"`
	Outputs        []nbOutput `json:"outputs,omitempty"`
}

type nbLanguageInfo struct {
	FileExtension *string `json:"file_extension,omitempty"`
}

type nbKernelSpec struct {
	DisplayName *string `json:"display_name,omitempty"`
	Language    *string `json:"language,omitempty"`
	Name        *string `json:"name,omitempty"`
}

type nbMetadata struct {
	LanguageInfo nbLanguageInfo `json:"language_info"`
	KernelSpec   nbKernelSpec   `json:"kernelspec"`
}

type nbNotebook struct {
	Cells         []nbCell   `json:"cells"`
	Metadata      nbMetadata `json:"metadata"`
	NBFormat      int        `json:"nbformat"`
	NBFormatMinor int        `json:"nbformat_minor"`
}

// --- ANSI escape code stripping ---

var ansiEscapeRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripANSI(s string) string {
	return ansiEscapeRegex.ReplaceAllString(s, "")
}

// --- HTML helpers ---

var nbSanitizerPolicy = func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowDataURIImages()
	return p
}()

func nbSanitizeHTML(htmlString string) template.HTML {
	return template.HTML(nbSanitizerPolicy.Sanitize(htmlString)) //nolint:gosec
}

func nbEscapeHTML(s string) template.HTML {
	return template.HTML(html.EscapeString(s)) //nolint:gosec
}

// --- Rendering ---

// nbRenderMarkdown converts Markdown to HTML using the shared goldmark parser.
func nbRenderMarkdown(markdown string) template.HTML {
	var buf bytes.Buffer
	if err := markdownParser.Convert([]byte(markdown), &buf); err != nil {
		return template.HTML(html.EscapeString(markdown)) //nolint:gosec
	}
	return template.HTML(buf.String()) //nolint:gosec
}

// nbRenderSourceCode converts source code to syntax-highlighted HTML using Chroma v2.
func nbRenderSourceCode(source string, languageID string) (template.HTML, error) {
	buf := new(bytes.Buffer)

	var l chroma.Lexer
	if languageID != "" {
		l = lexers.Get(languageID)
	}
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	formatter := htmlFormatter.New(htmlFormatter.WithClasses(true))

	iterator, err := l.Tokenise(nil, source)
	if err != nil {
		return "", fmt.Errorf("tokenization error: %w", err)
	}

	if err = formatter.Format(buf, styles.GitHub, iterator); err != nil {
		return "", fmt.Errorf("formatting error: %w", err)
	}

	return template.HTML(buf.String()), nil //nolint:gosec
}

// --- Output converters ---

func nbConvertDataOutput(output nbOutput) template.HTML {
	switch {
	case output.Data.TextHTML != nil:
		htmlString := strings.Join(output.Data.TextHTML, "")
		// Remove unnecessary wrapper <div>
		if strings.HasPrefix(htmlString, "<div>") && strings.HasSuffix(htmlString, "</div>") {
			htmlString = htmlString[5 : len(htmlString)-6]
		}
		return nbSanitizeHTML(htmlString)
	case output.Data.ImagePNG != nil:
		s := fmt.Sprintf(`<img src="data:image/png;base64,%s">`, html.EscapeString(*output.Data.ImagePNG))
		return nbSanitizeHTML(s)
	case output.Data.ImageJPEG != nil:
		s := fmt.Sprintf(`<img src="data:image/jpeg;base64,%s">`, html.EscapeString(*output.Data.ImageJPEG))
		return nbSanitizeHTML(s)
	case output.Data.ImageSVGXML != nil:
		return nbSanitizeHTML(strings.Join(output.Data.ImageSVGXML, ""))
	case output.Data.TextMarkdown != nil:
		return nbRenderMarkdown(strings.Join(output.Data.TextMarkdown, ""))
	case output.Data.TextPlain != nil:
		return "<pre>" + nbEscapeHTML(strings.Join(output.Data.TextPlain, "")) + "</pre>"
	default:
		return ""
	}
}

func nbConvertErrorOutput(output nbOutput) template.HTML {
	if output.Traceback == nil {
		return "<pre>An unknown error occurred</pre>"
	}
	lines := make([]string, len(output.Traceback))
	for i, line := range output.Traceback {
		lines[i] = stripANSI(line)
	}
	return "<pre>" + nbEscapeHTML(strings.Join(lines, "\n")) + "</pre>"
}

func nbConvertStreamOutput(output nbOutput) template.HTML {
	if output.Text == nil {
		return ""
	}
	return "<pre>" + nbEscapeHTML(strings.Join(output.Text, "")) + "</pre>"
}

// --- Cell converters ---

func nbConvertInput(languageID string, cell nbCell) template.HTML {
	switch cell.CellType {
	case "markdown":
		return nbRenderMarkdown(strings.Join(cell.Source, ""))
	case "code":
		source := strings.Join(cell.Source, "")
		h, err := nbRenderSourceCode(source, languageID)
		if err != nil {
			return "<pre>" + nbEscapeHTML(source) + "</pre>"
		}
		return h
	case "raw":
		return "<pre>" + nbEscapeHTML(strings.Join(cell.Source, "")) + "</pre>"
	default:
		return ""
	}
}

func nbConvertOutput(output nbOutput) template.HTML {
	switch output.OutputType {
	case "display_data", "execute_result":
		return nbConvertDataOutput(output)
	case "error":
		return nbConvertErrorOutput(output)
	case "stream":
		return nbConvertStreamOutput(output)
	default:
		return ""
	}
}

func nbConvertPrompt(executionCount *int) template.HTML {
	if executionCount == nil {
		return ""
	}
	return template.HTML(fmt.Sprintf("[%d]:", *executionCount)) //nolint:gosec
}

// --- Notebook template ---

var notebookTemplate = template.Must(template.New("notebook").Funcs(template.FuncMap{
	"convertPrompt": nbConvertPrompt,
	"convertInput":  nbConvertInput,
	"convertOutput": nbConvertOutput,
	"getCellClasses": func(cell nbCell) string {
		return "cell cell-" + cell.CellType
	},
	"getOutputClasses": func(output nbOutput) string {
		return "output output-" + strings.ReplaceAll(output.OutputType, "_", "-")
	},
}).Parse(`
		<div class="notebook">
			{{ $languageID := .languageID }}
			{{ range .notebook.Cells }}
				<div class="{{ . | getCellClasses }}">
					<div class="input-wrapper">
						<div class="input-prompt">
							{{ .ExecutionCount | convertPrompt }}
						</div>
						<div class="input">
							{{ . | convertInput $languageID }}
						</div>
					</div>
					{{ range .Outputs }}
						<div class="output-wrapper">
							<div class="output-prompt">
								{{ .ExecutionCount | convertPrompt }}
							</div>
							<div class="{{ . | getOutputClasses }}">
								{{ . | convertOutput }}
							</div>
						</div>
					{{ end }}
				</div>
			{{ end }}
		</div>
	`))

// --- Entry point ---

// convertNotebookString converts a Jupyter Notebook JSON string to HTML.
func convertNotebookString(notebookString string) (string, error) {
	var nb nbNotebook
	if err := json.Unmarshal([]byte(notebookString), &nb); err != nil {
		return "", fmt.Errorf("could not parse Jupyter Notebook JSON: %w", err)
	}

	if nb.NBFormat < 4 {
		return "", fmt.Errorf(
			"unsupported Jupyter Notebook format version %d (version 4+ required)",
			nb.NBFormat,
		)
	}

	// Detect the programming language from notebook metadata
	languageID := ""
	if ext := nb.Metadata.LanguageInfo.FileExtension; ext != nil {
		languageID = (*ext)[1:] // strip leading dot
	} else if lang := nb.Metadata.KernelSpec.Language; lang != nil {
		languageID = *lang
	} else if name := nb.Metadata.KernelSpec.Name; name != nil {
		languageID = *name
	}

	var buf bytes.Buffer
	templateVars := map[string]interface{}{
		"languageID": languageID,
		"notebook":   nb,
	}
	if err := notebookTemplate.Execute(&buf, templateVars); err != nil {
		return "", fmt.Errorf("could not render notebook template: %w", err)
	}

	return buf.String(), nil
}
