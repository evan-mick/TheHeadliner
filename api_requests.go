/// FOR ANY EXTERNAL API REQUESTS (not database stuff)
/// rn just gemini

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const prompt = `Act as a news summarizer 
and create a concise summary of the news article I provide. 
The summary can be up to 4 sentences in length, 
expressing the key points in paragraph form without line breaks. 
Include concepts written in the original article without adding your interpretations 
and include as many specific details from the article as possible. 
Now, please summarize this article: `

// Thank you Gemini for the types <3
type GeminiResponse struct {
	Candidates     []Candidate    `json:"candidates"`
	UsageMetadata  UsageMetadata  `json:"usageMetadata"`
	PromptFeedback PromptFeedback `json:"promptFeedback"`
}

type Candidate struct {
	Content       Content        `json:"content"`
	FinishReason  string         `json:"finishReason"`
	Index         int            `json:"index"`
	SafetyRatings []SafetyRating `json:"safetyRatings"`
}

type Content struct {
	Parts []Part `json:"parts"`
	Role  string `json:"role"`
}

type Part struct {
	Text string `json:"text"` // Gemini returns text in this field
}

type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

type SafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

type PromptFeedback struct {
	SafetyRatings []SafetyRating `json:"safetyRatings"`
}

func googleRequest(toSummarize string) (string, error) {

	google_key := os.Getenv("GOOGLE_AI_KEY")

	if google_key == "" {
		return "", fmt.Errorf("GOOGLE_AI_KEY environment variable is not set")
	}

	message := prompt + toSummarize

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{
						"text": message,
					},
				},
			},
		},
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshalling request body:", err)
		return "", err
	}

	// Define cURL equivalent URL
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-3-flash-preview:generateContent?key=" + google_key

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send HTTP
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)

		return "", err

	}
	defer resp.Body.Close()

	/*jsonStr := `{"candidates":
	[{"content":
		{"parts":
			[{"text":"The Tropicana, a historic Las Vegas Strip resort, will close on Tuesday after 67 years. Its demolition is scheduled for October to make way for a ballpark for Major League Baseball's Oakland A's.\n\nOnce Las Vegas' premier resort, the Tropicana hosted renowned shows like the Folies Bergère and welcomed famous guests like James Bond. However, in recent decades, it has fallen out of favor due to changing tourist preferences and competition from newer, more lucrative resorts.\n\nThe Tropicana's glamorous past will come to an end when it closes its doors, marking the departure of one of the Strip's last remaining resorts from the 1950s.\n\nThe new ballpark, expected to cost $1.5 billion, will feature a distinctive design inspired by traditional baseball pennants. The A's are set to move to Las Vegas in 2028, becoming the state's first major league team.\n\nWhile the Tropicana's closure signals the evolving nature of Las Vegas, it also highlights the city's continued growth and appeal as a major entertainment destination.\n\nThe A's' move to Las Vegas underscores the city's transformation into a thriving sports hub, attracting major events like the Super Bowl and Formula One races.\n\nThe Tropicana's legacy will live on through memories of its glamorous past and its role in the history of Las Vegas."}
			]
		},
		"role":"model",
		"finishReason":"STOP",
		"index":0,
		"safetyRatings":
		[{"category":"HARM_CATEGORY_SEXUALLY_EXPLICIT","probability":"NEGLIGIBLE"},{"category":"HARM_CATEGORY_HATE_SPEECH","probability":"NEGLIGIBLE"},{"category":"HARM_CATEGORY_HARASSMENT","probability":"NEGLIGIBLE"},{"category":"HARM_CATEGORY_DANGEROUS_CONTENT","probability":"NEGLIGIBLE"}]}],"promptFeedback":{"safetyRatings":[{"category":"HARM_CATEGORY_SEXUALLY_EXPLICIT","probability":"NEGLIGIBLE"},{"category":"HARM_CATEGORY_HATE_SPEECH","probability":"NEGLIGIBLE"},{"category":"HARM_CATEGORY_HARASSMENT","probability":"NEGLIGIBLE"},{"category":"HARM_CATEGORY_DANGEROUS_CONTENT","probability":"NEGLIGIBLE"}]}}`
	*/
	// Read response
	// var result map[string]interface{}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		// This will print the actual error message from Google (e.g., "API Key not found", "Model not found")
		return "", fmt.Errorf("google API error (Status %d): %s", resp.StatusCode, string(body))
	}

	var input GeminiResponse
	// err = json.NewDecoder(resp.Body).Decode(&result)
	err = json.Unmarshal([]byte(body), &input)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return "", err
	}

	if len(input.Candidates) > 0 && len(input.Candidates[0].Content.Parts) > 0 {
		//fmt.Println("-------------------- SUMMARY -------------------------")
		//fmt.Println(input.Candidates[0].Content.Parts[0].Text)
		return input.Candidates[0].Content.Parts[0].Text, nil
	}
	fmt.Println(input)
	//fmt.Println("NO CONTENT FOUND")
	return "", fmt.Errorf("No content found, Google API issue ")
}
