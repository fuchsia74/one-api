# OpenAI Function calling

- <https://platform.openai.com/docs/guides/function-calling>

## ChatCompletion API

Enable models to fetch data and take actions.

**Function calling** provides a powerful and flexible way for OpenAI models to interface with your code or external services. This guide will explain how to connect the models to your own custom code to fetch data or take action.

#### Get weather

Function calling example with get_weather function

```bash
curl https://api.openai.com/v1/chat/completions \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $OPENAI_API_KEY" \
-d '{
    "model": "gpt-4.1",
    "messages": [
        {
            "role": "user",
            "content": "What is the weather like in Paris today?"
        }
    ],
    "tools": [
        {
            "type": "function",
            "function": {
                "name": "get_weather",
                "description": "Get current temperature for a given location.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "location": {
                            "type": "string",
                            "description": "City and country e.g. Bogotá, Colombia"
                        }
                    },
                    "required": [
                        "location"
                    ],
                    "additionalProperties": false
                },
                "strict": true
            }
        }
    ]
}'
```

Output

```json
[
  {
    "id": "call_12345xyz",
    "type": "function",
    "function": {
      "name": "get_weather",
      "arguments": "{\"location\":\"Paris, France\"}"
    }
  }
]
```

#### Send email

Function calling example with send_email function

```bash
curl https://api.openai.com/v1/chat/completions \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $OPENAI_API_KEY" \
-d '{
    "model": "gpt-4.1",
    "messages": [
        {
            "role": "user",
            "content": "Can you send an email to ilan@example.com and katia@example.com saying hi?"
        }
    ],
    "tools": [
        {
            "type": "function",
            "function": {
                "name": "send_email",
                "description": "Send an email to a given recipient with a subject and message.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "to": {
                            "type": "string",
                            "description": "The recipient email address."
                        },
                        "subject": {
                            "type": "string",
                            "description": "Email subject line."
                        },
                        "body": {
                            "type": "string",
                            "description": "Body of the email message."
                        }
                    },
                    "required": [
                        "to",
                        "subject",
                        "body"
                    ],
                    "additionalProperties": false
                },
                "strict": true
            }
        }
    ]
}'
```

Output

```json
[
  {
    "id": "call_9876abc",
    "type": "function",
    "function": {
      "name": "send_email",
      "arguments": "{\"to\":\"ilan@example.com\",\"subject\":\"Hello!\",\"body\":\"Just wanted to say hi\"}"
    }
  },
  {
    "id": "call_9876abc",
    "type": "function",
    "function": {
      "name": "send_email",
      "arguments": "{\"to\":\"katia@example.com\",\"subject\":\"Hello!\",\"body\":\"Just wanted to say hi\"}"
    }
  }
]
```

#### Search knowledge base

Function calling example with search_knowledge_base function

```bash
curl https://api.openai.com/v1/chat/completions \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $OPENAI_API_KEY" \
-d '{
    "model": "gpt-4.1",
    "messages": [
        {
            "role": "user",
            "content": "Can you find information about ChatGPT in the AI knowledge base?"
        }
    ],
    "tools": [
        {
            "type": "function",
            "function": {
                "name": "search_knowledge_base",
                "description": "Query a knowledge base to retrieve relevant info on a topic.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "query": {
                            "type": "string",
                            "description": "The user question or search query."
                        },
                        "options": {
                            "type": "object",
                            "properties": {
                                "num_results": {
                                    "type": "number",
                                    "description": "Number of top results to return."
                                },
                                "domain_filter": {
                                    "type": [
                                        "string",
                                        "null"
                                    ],
                                    "description": "Optional domain to narrow the search (e.g. 'finance', 'medical'). Pass null if not needed."
                                },
                                "sort_by": {
                                    "type": [
                                        "string",
                                        "null"
                                    ],
                                    "enum": [
                                        "relevance",
                                        "date",
                                        "popularity",
                                        "alphabetical"
                                    ],
                                    "description": "How to sort results. Pass null if not needed."
                                }
                            },
                            "required": [
                                "num_results",
                                "domain_filter",
                                "sort_by"
                            ],
                            "additionalProperties": false
                        }
                    },
                    "required": [
                        "query",
                        "options"
                    ],
                    "additionalProperties": false
                },
                "strict": true
            }
        }
    ]
}'
```

Output

```json
[
  {
    "id": "call_4567xyz",
    "type": "function",
    "function": {
      "name": "search_knowledge_base",
      "arguments": "{\"query\":\"What is ChatGPT?\",\"options\":{\"num_results\":3,\"domain_filter\":null,\"sort_by\":\"relevance\"}}"
    }
  }
]
```

Experiment with function calling and [generate function schemas](/docs/guides/prompt-generation) in the [Playground](/playground)!

### Overview

You can give the model access to your own custom code through **function calling**. Based on the system prompt and messages, the model may decide to call these functions — **instead of (or in addition to) generating text or audio**.

You'll then execute the function code, send back the results, and the model will incorporate them into its final response.

![Function Calling Diagram Steps](https://cdn.openai.com/API/docs/images/function-calling-diagram-steps.png)

Function calling has two primary use cases:

|               |                                                                                                                                                                                          |
| ------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Fetching Data | Retrieve up-to-date information to incorporate into the model's response (RAG). Useful for searching knowledge bases and retrieving specific data from APIs (e.g. current weather data). |
| Taking Action | Perform actions like submitting a form, calling APIs, modifying application state (UI/frontend or backend), or taking agentic workflow actions (like handing off the conversation).      |

### Sample function

Let's look at the steps to allow a model to use a real `get_weather` function defined below:

Sample get_weather function implemented in your codebase

```javascript
async function getWeather(latitude, longitude) {
  const response = await fetch(
    `https://api.open-meteo.com/v1/forecast?latitude=${latitude}&longitude=${longitude}&current=temperature_2m,wind_speed_10m&hourly=temperature_2m,relative_humidity_2m,wind_speed_10m`
  );
  const data = await response.json();
  return data.current.temperature_2m;
}
```

Unlike the diagram earlier, this function expects precise `latitude` and `longitude` instead of a general `location` parameter. (However, our models can automatically determine the coordinates for many locations!)

### Function calling steps

- **Call model with [functions defined](#defining-functions)** – along with your system and user messages.

Step 1: Call model with get_weather tool defined

```javascript
import { OpenAI } from "openai";

const openai = new OpenAI();

const tools = [
  {
    type: "function",
    function: {
      name: "get_weather",
      description:
        "Get current temperature for provided coordinates in celsius.",
      parameters: {
        type: "object",
        properties: {
          latitude: { type: "number" },
          longitude: { type: "number" },
        },
        required: ["latitude", "longitude"],
        additionalProperties: false,
      },
      strict: true,
    },
  },
];

const messages = [
  {
    role: "user",
    content: "What's the weather like in Paris today?",
  },
];

const completion = await openai.chat.completions.create({
  model: "gpt-4.1",
  messages,
  tools,
  store: true,
});
```

- **Model decides to call function(s)** – model returns the **name** and **input arguments**.

completion.choices\[0\].message.tool_calls

```json
[
  {
    "id": "call_12345xyz",
    "type": "function",
    "function": {
      "name": "get_weather",
      "arguments": "{\"latitude\":48.8566,\"longitude\":2.3522}"
    }
  }
]
```

- **Execute function code** – parse the model's response and [handle function calls](#handling-function-calls).

Step 3: Execute get_weather function

```javascript
const toolCall = completion.choices[0].message.tool_calls[0];
const args = JSON.parse(toolCall.function.arguments);

const result = await getWeather(args.latitude, args.longitude);
```

- **Supply model with results** – so it can incorporate them into its final response.

Step 4: Supply result and call model again

```javascript
messages.push(completion.choices[0].message); // append model's function call message
messages.push({
  // append result message
  role: "tool",
  tool_call_id: toolCall.id,
  content: result.toString(),
});

const completion2 = await openai.chat.completions.create({
  model: "gpt-4.1",
  messages,
  tools,
  store: true,
});

console.log(completion2.choices[0].message.content);
```

- **Model responds** – incorporating the result in its output.

completion_2.choices\[0\].message.content

```json
"The current temperature in Paris is 14°C (57.2°F)."
```

### Defining functions

Functions can be set in the `tools` parameter of each API request inside a `function` object.

A function is defined by its schema, which informs the model what it does and what input arguments it expects. It comprises the following fields:

| Field       | Description                                         |
| ----------- | --------------------------------------------------- |
| name        | The function's name (e.g. get_weather)              |
| description | Details on when and how to use the function         |
| parameters  | JSON schema defining the function's input arguments |

Take a look at this example or generate your own below (or in our [Playground](/playground)).

```json
{
  "type": "function",
  "function": {
    "name": "get_weather",
    "description": "Retrieves current weather for the given location.",
    "parameters": {
      "type": "object",
      "properties": {
        "location": {
          "type": "string",
          "description": "City and country e.g. Bogotá, Colombia"
        },
        "units": {
          "type": "string",
          "enum": ["celsius", "fahrenheit"],
          "description": "Units the temperature will be returned in."
        }
      },
      "required": ["location", "units"],
      "additionalProperties": false
    },
    "strict": true
  }
}
```

Because the `parameters` are defined by a [JSON schema](https://json-schema.org/), you can leverage many of its rich features like property types, enums, descriptions, nested objects, and, recursive objects.

(Optional) Function calling wth pydantic and zod

While we encourage you to define your function schemas directly, our SDKs have helpers to convert `pydantic` and `zod` objects into schemas. Not all `pydantic` and `zod` features are supported.

Define objects to represent function schema

```javascript
import OpenAI from "openai";
import { z } from "zod";
import { zodFunction } from "openai/helpers/zod";

const openai = new OpenAI();

const GetWeatherParameters = z.object({
  location: z.string().describe("City and country e.g. Bogotá, Colombia"),
});

const tools = [
  zodFunction({ name: "getWeather", parameters: GetWeatherParameters }),
];

const messages = [
  { role: "user", content: "What's the weather like in Paris today?" },
];

const response = await openai.chat.completions.create({
  model: "gpt-4.1",
  messages,
  tools,
  store: true,
});

console.log(response.choices[0].message.tool_calls);
```

### Best practices for defining functions

1.  **Write clear and detailed function names, parameter descriptions, and instructions.**

    - **Explicitly describe the purpose of the function and each parameter** (and its format), and what the output represents.
    - **Use the system prompt to describe when (and when not) to use each function.** Generally, tell the model _exactly_ what to do.
    - **Include examples and edge cases**, especially to rectify any recurring failures. (**Note:** Adding examples may hurt performance for [reasoning models](/docs/guides/reasoning).)

2.  **Apply software engineering best practices.**

    - **Make the functions obvious and intuitive**. ([principle of least surprise](https://en.wikipedia.org/wiki/Principle_of_least_astonishment))
    - **Use enums** and object structure to make invalid states unrepresentable. (e.g. `toggle_light(on: bool, off: bool)` allows for invalid calls)
    - **Pass the intern test.** Can an intern/human correctly use the function given nothing but what you gave the model? (If not, what questions do they ask you? Add the answers to the prompt.)

3.  **Offload the burden from the model and use code where possible.**

    - **Don't make the model fill arguments you already know.** For example, if you already have an `order_id` based on a previous menu, don't have an `order_id` param – instead, have no params `submit_refund()` and pass the `order_id` with code.
    - **Combine functions that are always called in sequence.** For example, if you always call `mark_location()` after `query_location()`, just move the marking logic into the query function call.

4.  **Keep the number of functions small for higher accuracy.**

    - **Evaluate your performance** with different numbers of functions.
    - **Aim for fewer than 20 functions** at any one time, though this is just a soft suggestion.

5.  **Leverage OpenAI resources.**

    - **Generate and iterate on function schemas** in the [Playground](/playground).
    - **Consider [fine-tuning](https://platform.openai.com/docs/guides/fine-tuning) to increase function calling accuracy** for large numbers of functions or difficult tasks. ([cookbook](https://cookbook.openai.com/examples/fine_tuning_for_function_calling))

### Token Usage

Under the hood, functions are injected into the system message in a syntax the model has been trained on. This means functions count against the model's context limit and are billed as input tokens. If you run into token limits, we suggest limiting the number of functions or the length of the descriptions you provide for function parameters.

It is also possible to use [fine-tuning](/docs/guides/fine-tuning#fine-tuning-examples) to reduce the number of tokens used if you have many functions defined in your tools specification.

### Handling function calls

When the model calls a function, you must execute it and return the result. Since model responses can include zero, one, or multiple calls, it is best practice to assume there are several.

The response has an array of `tool_calls`, each with an `id` (used later to submit the function result) and a `function` containing a `name` and JSON-encoded `arguments`.

Sample response with multiple function calls

```json
[
  {
    "id": "call_12345xyz",
    "type": "function",
    "function": {
      "name": "get_weather",
      "arguments": "{\"location\":\"Paris, France\"}"
    }
  },
  {
    "id": "call_67890abc",
    "type": "function",
    "function": {
      "name": "get_weather",
      "arguments": "{\"location\":\"Bogotá, Colombia\"}"
    }
  },
  {
    "id": "call_99999def",
    "type": "function",
    "function": {
      "name": "send_email",
      "arguments": "{\"to\":\"bob@email.com\",\"body\":\"Hi bob\"}"
    }
  }
]
```

Execute function calls and append results

```javascript
for (const toolCall of completion.choices[0].message.tool_calls) {
  const name = toolCall.function.name;
  const args = JSON.parse(toolCall.function.arguments);

  const result = callFunction(name, args);
  messages.push({
    role: "tool",
    tool_call_id: toolCall.id,
    content: result.toString(),
  });
}
```

In the example above, we have a hypothetical `call_function` to route each call. Here’s a possible implementation:

Execute function calls and append results

```javascript
const callFunction = async (name, args) => {
  if (name === "get_weather") {
    return getWeather(args.latitude, args.longitude);
  }
  if (name === "send_email") {
    return sendEmail(args.to, args.body);
  }
};
```

### Formatting results

A result must be a string, but the format is up to you (JSON, error codes, plain text, etc.). The model will interpret that string as needed.

If your function has no return value (e.g. `send_email`), simply return a string to indicate success or failure. (e.g. `"success"`)

### Incorporating results into response

After appending the results to your `messages`, you can send them back to the model to get a final response.

Send results back to model

```javascript
const completion = await openai.chat.completions.create({
  model: "gpt-4.1",
  messages,
  tools,
  store: true,
});
```

Final response

```json
"It's about 15°C in Paris, 18°C in Bogotá, and I've sent that email to Bob."
```

### Additional configurations

### Tool choice

By default the model will determine when and how many tools to use. You can force specific behavior with the `tool_choice` parameter.

1.  **Auto:** (_Default_) Call zero, one, or multiple functions. `tool_choice: "auto"`
2.  **Required:** Call one or more functions. `tool_choice: "required"`

3.  **Forced Function:** Call exactly one specific function. `tool_choice: {"type": "function", "function": {"name": "get_weather"}}`

![Function Calling Diagram Steps](https://cdn.openai.com/API/docs/images/function-calling-diagram-tool-choice.png)

You can also set `tool_choice` to `"none"` to imitate the behavior of passing no functions.

### Parallel function calling

The model may choose to call multiple functions in a single turn. You can prevent this by setting `parallel_tool_calls` to `false`, which ensures exactly zero or one tool is called.

**Note:** Currently, if you are using a fine tuned model and the model calls multiple functions in one turn then [strict mode](#strict-mode) will be disabled for those calls.

**Note for `gpt-4.1-nano-2025-04-14`:** This snapshot of `gpt-4.1-nano` can sometimes include multiple tools calls for the same tool if parallel tool calls are enabled. It is recommended to disable this feature when using this nano snapshot.

### Strict mode

Setting `strict` to `true` will ensure function calls reliably adhere to the function schema, instead of being best effort. We recommend always enabling strict mode.

Under the hood, strict mode works by leveraging our [structured outputs](/docs/guides/structured-outputs) feature and therefore introduces a couple requirements:

1.  `additionalProperties` must be set to `false` for each object in the `parameters`.
2.  All fields in `properties` must be marked as `required`.

You can denote optional fields by adding `null` as a `type` option (see example below).

Strict mode enabled

```json
{
  "type": "function",
  "function": {
    "name": "get_weather",
    "description": "Retrieves current weather for the given location.",
    "strict": true,
    "parameters": {
      "type": "object",
      "properties": {
        "location": {
          "type": "string",
          "description": "City and country e.g. Bogotá, Colombia"
        },
        "units": {
          "type": ["string", "null"],
          "enum": ["celsius", "fahrenheit"],
          "description": "Units the temperature will be returned in."
        }
      },
      "required": ["location", "units"],
      "additionalProperties": false
    }
  }
}
```

Strict mode disabled

```json
{
  "type": "function",
  "function": {
    "name": "get_weather",
    "description": "Retrieves current weather for the given location.",
    "parameters": {
      "type": "object",
      "properties": {
        "location": {
          "type": "string",
          "description": "City and country e.g. Bogotá, Colombia"
        },
        "units": {
          "type": "string",
          "enum": ["celsius", "fahrenheit"],
          "description": "Units the temperature will be returned in."
        }
      },
      "required": ["location"]
    }
  }
}
```

All schemas generated in the [playground](/playground) have strict mode enabled.

While we recommend you enable strict mode, it has a few limitations:

1.  Some features of JSON schema are not supported. (See [supported schemas](/docs/guides/structured-outputs?context=with_parse#supported-schemas).)

Specifically for fine tuned models:

1.  Schemas undergo additional processing on the first request (and are then cached). If your schemas vary from request to request, this may result in higher latencies.
2.  Schemas are cached for performance, and are not eligible for [zero data retention](/docs/models#how-we-use-your-data).

### Streaming

Streaming can be used to surface progress by showing which function is called as the model fills its arguments, and even displaying the arguments in real time.

Streaming function calls is very similar to streaming regular responses: you set `stream` to `true` and get chunks with `delta` objects.

Streaming function calls

```javascript
import { OpenAI } from "openai";

const openai = new OpenAI();

const tools = [
  {
    type: "function",
    function: {
      name: "get_weather",
      description: "Get current temperature for a given location.",
      parameters: {
        type: "object",
        properties: {
          location: {
            type: "string",
            description: "City and country e.g. Bogotá, Colombia",
          },
        },
        required: ["location"],
        additionalProperties: false,
      },
      strict: true,
    },
  },
];

const stream = await openai.chat.completions.create({
  model: "gpt-4.1",
  messages: [
    { role: "user", content: "What's the weather like in Paris today?" },
  ],
  tools,
  stream: true,
  store: true,
});

for await (const chunk of stream) {
  const delta = chunk.choices[0].delta;
  console.log(delta.tool_calls);
}
```

Output delta.tool_calls

```json
[{"index": 0, "id": "call_DdmO9pD3xa9XTPNJ32zg2hcA", "function": {"arguments": "", "name": "get_weather"}, "type": "function"}]
[{"index": 0, "id": null, "function": {"arguments": "{\"", "name": null}, "type": null}]
[{"index": 0, "id": null, "function": {"arguments": "location", "name": null}, "type": null}]
[{"index": 0, "id": null, "function": {"arguments": "\":\"", "name": null}, "type": null}]
[{"index": 0, "id": null, "function": {"arguments": "Paris", "name": null}, "type": null}]
[{"index": 0, "id": null, "function": {"arguments": ",", "name": null}, "type": null}]
[{"index": 0, "id": null, "function": {"arguments": " France", "name": null}, "type": null}]
[{"index": 0, "id": null, "function": {"arguments": "\"}", "name": null}, "type": null}]
null
```

Instead of aggregating chunks into a single `content` string, however, you're aggregating chunks into an encoded `arguments` JSON object.

When the model calls one or more functions the `tool_calls` field of each `delta` will be populated. Each `tool_call` contains the following fields:

| Field    | Description                                            |
| -------- | ------------------------------------------------------ |
| index    | Identifies which function call the delta is for        |
| id       | Tool call id.                                          |
| function | Function call delta (name and arguments)               |
| type     | Type of tool_call (always function for function calls) |

Many of these fields are only set for the first `delta` of each tool call, like `id`, `function.name`, and `type`.

Below is a code snippet demonstrating how to aggregate the `delta`s into a final `tool_calls` object.

Accumulating tool_call deltas

```javascript
const finalToolCalls = {};

for await (const chunk of stream) {
  const toolCalls = chunk.choices[0].delta.tool_calls || [];
  for (const toolCall of toolCalls) {
    const { index } = toolCall;

    if (!finalToolCalls[index]) {
      finalToolCalls[index] = toolCall;
    }

    finalToolCalls[index].function.arguments += toolCall.function.arguments;
  }
}
```

Accumulated final_tool_calls\[0\]

```json
{
  "index": 0,
  "id": "call_RzfkBpJgzeR0S242qfvjadNe",
  "function": {
    "name": "get_weather",
    "arguments": "{\"location\":\"Paris, France\"}"
  }
}
```

### ChatCompletion Web Search

Allow models to search the web for the latest information before generating a response.

Using the [Chat Completions API](/docs/api-reference/chat), you can directly access the fine-tuned models and tool used by [Search in ChatGPT](https://openai.com/index/introducing-chatgpt-search/).

When using Chat Completions, the model always retrieves information from the web before responding to your query. To use `web_search_preview` as a tool that models like `gpt-4o` and `gpt-4o-mini` invoke only when necessary, switch to using the [Responses API](/docs/guides/tools-web-search?api-mode=responses).

Currently, you need to use one of these models to use web search in Chat Completions:

- `gpt-4o-search-preview`
- `gpt-4o-mini-search-preview`

Web search parameter example

```javascript
import OpenAI from "openai";
const client = new OpenAI();

const completion = await client.chat.completions.create({
  model: "gpt-4o-search-preview",
  web_search_options: {},
  messages: [
    {
      role: "user",
      content: "What was a positive news story from today?",
    },
  ],
});

console.log(completion.choices[0].message.content);
```

```python
from openai import OpenAI
client = OpenAI()

completion = client.chat.completions.create(
    model="gpt-4o-search-preview",
    web_search_options={},
    messages=[
        {
            "role": "user",
            "content": "What was a positive news story from today?",
        }
    ],
)

print(completion.choices[0].message.content)
```

```bash
curl -X POST "https://api.openai.com/v1/chat/completions" \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    -H "Content-type: application/json" \
    -d '{
        "model": "gpt-4o-search-preview",
        "web_search_options": {},
        "messages": [{
            "role": "user",
            "content": "What was a positive news story from today?"
        }]
    }'
```

## Output and citations

The API response item in the `choices` array will include:

- `message.content` with the text result from the model, inclusive of any inline citations
- `annotations` with a list of cited URLs

By default, the model's response will include inline citations for URLs found in the web search results. In addition to this, the `url_citation` annotation object will contain the URL and title of the cited source, as well as the start and end index characters in the model's response where those sources were used.

When displaying web results or information contained in web results to end users, inline citations must be made clearly visible and clickable in your user interface.

```json
[
  {
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "the model response is here...",
      "refusal": null,
      "annotations": [
        {
          "type": "url_citation",
          "url_citation": {
            "end_index": 985,
            "start_index": 764,
            "title": "Page title...",
            "url": "https://..."
          }
        }
      ]
    },
    "finish_reason": "stop"
  }
]
```

## Domain filtering

Domain filtering in web search lets you limit results to a specific set of domains. With the `filters` parameter you can set an allow-list of up to 20 URLs. When formatting URLs, omit the HTTP or HTTPS prefix. For example, use [`openai.com`](http://openai.com) instead of [`https://openai.com/`](https://openai.com/). This approach also includes subdomains in the search. Note that domain filtering is only available in the Responses API with the `web_search` tool.

## Sources

To view all URLs retrieved during a web search, use the `sources` field. Unlike inline citations, which show only the most relevant references, sources returns the complete list of URLs the model consulted when forming its response. The number of sources is often greater than the number of citations. Real-time third-party feeds are also surfaced here and are labeled as `oai-sports`, `oai-weather`, or `oai-finance`. The sources field is available with both the `web_search` and `web_search_preview` tools.

List sources

```bash
curl "https://api.openai.com/v1/responses" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $OPENAI_API_KEY" \
-d '{
  "model": "gpt-5",
  "reasoning": { "effort": "low" },
  "tools": [
    {
      "type": "web_search",
      "filters": {
        "allowed_domains": [
          "pubmed.ncbi.nlm.nih.gov",
          "clinicaltrials.gov",
          "www.who.int",
          "www.cdc.gov",
          "www.fda.gov"
        ]
      }
    }
  ],
  "tool_choice": "auto",
  "include": ["web_search_call.action.sources"],
  "input": "Please perform a web search on how semaglutide is used in the treatment of diabetes."
}'
```

```javascript
import OpenAI from "openai";
const client = new OpenAI();

const response = await client.responses.create({
  model: "gpt-5",
  reasoning: { effort: "low" },
  tools: [
    {
      type: "web_search",
      filters: {
        allowed_domains: [
          "pubmed.ncbi.nlm.nih.gov",
          "clinicaltrials.gov",
          "www.who.int",
          "www.cdc.gov",
          "www.fda.gov",
        ],
      },
    },
  ],
  tool_choice: "auto",
  include: ["web_search_call.action.sources"],
  input:
    "Please perform a web search on how semaglutide is used in the treatment of diabetes.",
});

console.log(response.output_text);
```

```python
from openai import OpenAI
client = OpenAI()

response = client.responses.create(
  model="gpt-5",
  reasoning={"effort": "low"},
  tools=[
      {
          "type": "web_search",
          "filters": {
              "allowed_domains": [
                  "pubmed.ncbi.nlm.nih.gov",
                  "clinicaltrials.gov",
                  "www.who.int",
                  "www.cdc.gov",
                  "www.fda.gov",
              ]
          },
      }
  ],
  tool_choice="auto",
  include=["web_search_call.action.sources"],
  input="Please perform a web search on how semaglutide is used in the treatment of diabetes.",
)

print(response.output_text)
```

## User location

To refine search results based on geography, you can specify an approximate user location using country, city, region, and/or timezone.

- The `city` and `region` fields are free text strings, like `Minneapolis` and `Minnesota` respectively.
- The `country` field is a two-letter [ISO country code](https://en.wikipedia.org/wiki/ISO_3166-1), like `US`.
- The `timezone` field is an [IANA timezone](https://timeapi.io/documentation/iana-timezones) like `America/Chicago`.

Note that user location is not supported for deep research models using web search.

Customizing user location

```python
from openai import OpenAI
client = OpenAI()

completion = client.chat.completions.create(
    model="gpt-4o-search-preview",
    web_search_options={
        "user_location": {
            "type": "approximate",
            "approximate": {
                "country": "GB",
                "city": "London",
                "region": "London",
            }
        },
    },
    messages=[{
        "role": "user",
        "content": "What are the best restaurants around Granary Square?",
    }],
)

print(completion.choices[0].message.content)
```

```javascript
import OpenAI from "openai";
const client = new OpenAI();

const completion = await client.chat.completions.create({
  model: "gpt-4o-search-preview",
  web_search_options: {
    user_location: {
      type: "approximate",
      approximate: {
        country: "GB",
        city: "London",
        region: "London",
      },
    },
  },
  messages: [
    {
      role: "user",
      content: "What are the best restaurants around Granary Square?",
    },
  ],
});
console.log(completion.choices[0].message.content);
```

```bash
curl -X POST "https://api.openai.com/v1/chat/completions" \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    -H "Content-type: application/json" \
    -d '{
        "model": "gpt-4o-search-preview",
        "web_search_options": {
            "user_location": {
                "type": "approximate",
                "approximate": {
                    "country": "GB",
                    "city": "London",
                    "region": "London"
                }
            }
        },
        "messages": [{
            "role": "user",
            "content": "What are the best restaurants around Granary Square?"
        }]
    }'
```

## API compatibility

Web search is available in the Responses API as the generally available version of the tool, `web_search`, as well as the earlier tool version, `web_search_preview`. To use web search in the Chat Completions API, use the specialized web search models `gpt-4o-search-preview` and `gpt-4o-mini-search-preview`.

## Limitations

- Web search is currently not supported in [`gpt-5`](/docs/models/gpt-5) with `minimal` reasoning, and [`gpt-4.1-nano`](/docs/models/gpt-4.1-nano).
- When used as a tool in the [Responses API](/docs/api-reference/responses), web search has the same tiered rate limits as the models above.
- Web search is limited to a context window size of 128000 (even with [`gpt-4.1`](/docs/models/gpt-4.1) and [`gpt-4.1-mini`](/docs/models/gpt-4.1-mini) models).

## Usage notes

||
|ResponsesChat CompletionsAssistants|Same as tiered rate limits for underlying model used with the tool.|PricingZDR and data residency|

## Response API

Enable models to fetch data and take actions.

**Function calling** provides a powerful and flexible way for OpenAI models to interface with your code or external services. This guide will explain how to connect the models to your own custom code to fetch data or take action.

#### Get weather

Function calling example with get_weather function

```bash
curl https://api.openai.com/v1/responses \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $OPENAI_API_KEY" \
-d '{
    "model": "gpt-4.1",
    "input": "What is the weather like in Paris today?",
    "tools": [
        {
            "type": "function",
            "name": "get_weather",
            "description": "Get current temperature for a given location.",
            "parameters": {
                "type": "object",
                "properties": {
                    "location": {
                        "type": "string",
                        "description": "City and country e.g. Bogotá, Colombia"
                    }
                },
                "required": [
                    "location"
                ],
                "additionalProperties": false
            }
        }
    ]
}'
```

Output

```json
[
  {
    "type": "function_call",
    "id": "fc_12345xyz",
    "call_id": "call_12345xyz",
    "name": "get_weather",
    "arguments": "{\"location\":\"Paris, France\"}"
  }
]
```

#### Send email

Function calling example with send_email function

```bash
curl https://api.openai.com/v1/responses \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $OPENAI_API_KEY" \
-d '{
    "model": "gpt-4.1",
    "input": "Can you send an email to ilan@example.com and katia@example.com saying hi?",
    "tools": [
        {
            "type": "function",
            "name": "send_email",
            "description": "Send an email to a given recipient with a subject and message.",
            "parameters": {
                "type": "object",
                "properties": {
                    "to": {
                        "type": "string",
                        "description": "The recipient email address."
                    },
                    "subject": {
                        "type": "string",
                        "description": "Email subject line."
                    },
                    "body": {
                        "type": "string",
                        "description": "Body of the email message."
                    }
                },
                "required": [
                    "to",
                    "subject",
                    "body"
                ],
                "additionalProperties": false
            }
        }
    ]
}'
```

Output

```json
[
  {
    "type": "function_call",
    "id": "fc_12345xyz",
    "call_id": "call_9876abc",
    "name": "send_email",
    "arguments": "{\"to\":\"ilan@example.com\",\"subject\":\"Hello!\",\"body\":\"Just wanted to say hi\"}"
  },
  {
    "type": "function_call",
    "id": "fc_12345xyz",
    "call_id": "call_9876abc",
    "name": "send_email",
    "arguments": "{\"to\":\"katia@example.com\",\"subject\":\"Hello!\",\"body\":\"Just wanted to say hi\"}"
  }
]
```

#### Search knowledge base

Function calling example with search_knowledge_base function

```bash
curl https://api.openai.com/v1/responses \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $OPENAI_API_KEY" \
-d '{
    "model": "gpt-4.1",
    "input": "Can you find information about ChatGPT in the AI knowledge base?",
    "tools": [
        {
            "type": "function",
            "name": "search_knowledge_base",
            "description": "Query a knowledge base to retrieve relevant info on a topic.",
            "parameters": {
                "type": "object",
                "properties": {
                    "query": {
                        "type": "string",
                        "description": "The user question or search query."
                    },
                    "options": {
                        "type": "object",
                        "properties": {
                            "num_results": {
                                "type": "number",
                                "description": "Number of top results to return."
                            },
                            "domain_filter": {
                                "type": [
                                    "string",
                                    "null"
                                ],
                                "description": "Optional domain to narrow the search (e.g. 'finance', 'medical'). Pass null if not needed."
                            },
                            "sort_by": {
                                "type": [
                                    "string",
                                    "null"
                                ],
                                "enum": [
                                    "relevance",
                                    "date",
                                    "popularity",
                                    "alphabetical"
                                ],
                                "description": "How to sort results. Pass null if not needed."
                            }
                        },
                        "required": [
                            "num_results",
                            "domain_filter",
                            "sort_by"
                        ],
                        "additionalProperties": false
                    }
                },
                "required": [
                    "query",
                    "options"
                ],
                "additionalProperties": false
            }
        }
    ]
}'
```

Output

```json
[
  {
    "type": "function_call",
    "id": "fc_12345xyz",
    "call_id": "call_4567xyz",
    "name": "search_knowledge_base",
    "arguments": "{\"query\":\"What is ChatGPT?\",\"options\":{\"num_results\":3,\"domain_filter\":null,\"sort_by\":\"relevance\"}}"
  }
]
```

Experiment with function calling and [generate function schemas](/docs/guides/prompt-generation) in the [Playground](/playground)!

### Overview

You can give the model access to your own custom code through **function calling**. Based on the system prompt and messages, the model may decide to call these functions — **instead of (or in addition to) generating text or audio**.

You'll then execute the function code, send back the results, and the model will incorporate them into its final response.

![Function Calling Diagram Steps](https://cdn.openai.com/API/docs/images/function-calling-diagram-steps.png)

Function calling has two primary use cases:

|               |                                                                                                                                                                                          |
| ------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Fetching Data | Retrieve up-to-date information to incorporate into the model's response (RAG). Useful for searching knowledge bases and retrieving specific data from APIs (e.g. current weather data). |
| Taking Action | Perform actions like submitting a form, calling APIs, modifying application state (UI/frontend or backend), or taking agentic workflow actions (like handing off the conversation).      |

### Sample function

Let's look at the steps to allow a model to use a real `get_weather` function defined below:

Sample get_weather function implemented in your codebase

```javascript
async function getWeather(latitude, longitude) {
  const response = await fetch(
    `https://api.open-meteo.com/v1/forecast?latitude=${latitude}&longitude=${longitude}&current=temperature_2m,wind_speed_10m&hourly=temperature_2m,relative_humidity_2m,wind_speed_10m`
  );
  const data = await response.json();
  return data.current.temperature_2m;
}
```

Unlike the diagram earlier, this function expects precise `latitude` and `longitude` instead of a general `location` parameter. (However, our models can automatically determine the coordinates for many locations!)

### Function calling steps

- **Call model with [functions defined](#defining-functions)** – along with your system and user messages.

Step 1: Call model with get_weather tool defined

```javascript
import { OpenAI } from "openai";

const openai = new OpenAI();

const tools = [
  {
    type: "function",
    name: "get_weather",
    description: "Get current temperature for provided coordinates in celsius.",
    parameters: {
      type: "object",
      properties: {
        latitude: { type: "number" },
        longitude: { type: "number" },
      },
      required: ["latitude", "longitude"],
      additionalProperties: false,
    },
    strict: true,
  },
];

const input = [
  {
    role: "user",
    content: "What's the weather like in Paris today?",
  },
];

const response = await openai.responses.create({
  model: "gpt-4.1",
  input,
  tools,
});
```

- **Model decides to call function(s)** – model returns the **name** and **input arguments**.

response.output

```json
[
  {
    "type": "function_call",
    "id": "fc_12345xyz",
    "call_id": "call_12345xyz",
    "name": "get_weather",
    "arguments": "{\"latitude\":48.8566,\"longitude\":2.3522}"
  }
]
```

- **Execute function code** – parse the model's response and [handle function calls](#handling-function-calls).

Step 3: Execute get_weather function

```javascript
const toolCall = response.output[0];
const args = JSON.parse(toolCall.arguments);

const result = await getWeather(args.latitude, args.longitude);
```

- **Supply model with results** – so it can incorporate them into its final response.

Step 4: Supply result and call model again

```javascript
input.push(toolCall); // append model's function call message
input.push({
  // append result message
  type: "function_call_output",
  call_id: toolCall.call_id,
  output: result.toString(),
});

const response2 = await openai.responses.create({
  model: "gpt-4.1",
  input,
  tools,
  store: true,
});

console.log(response2.output_text);
```

- **Model responds** – incorporating the result in its output.

response_2.output_text

```json
"The current temperature in Paris is 14°C (57.2°F)."
```

### Defining functions

Functions can be set in the `tools` parameter of each API request.

A function is defined by its schema, which informs the model what it does and what input arguments it expects. It comprises the following fields:

| Field       | Description                                          |
| ----------- | ---------------------------------------------------- |
| type        | This should always be function                       |
| name        | The function's name (e.g. get_weather)               |
| description | Details on when and how to use the function          |
| parameters  | JSON schema defining the function's input arguments  |
| strict      | Whether to enforce strict mode for the function call |

Take a look at this example or generate your own below (or in our [Playground](/playground)).

```json
{
  "type": "function",
  "name": "get_weather",
  "description": "Retrieves current weather for the given location.",
  "parameters": {
    "type": "object",
    "properties": {
      "location": {
        "type": "string",
        "description": "City and country e.g. Bogotá, Colombia"
      },
      "units": {
        "type": "string",
        "enum": ["celsius", "fahrenheit"],
        "description": "Units the temperature will be returned in."
      }
    },
    "required": ["location", "units"],
    "additionalProperties": false
  },
  "strict": true
}
```

Because the `parameters` are defined by a [JSON schema](https://json-schema.org/), you can leverage many of its rich features like property types, enums, descriptions, nested objects, and, recursive objects.

### Best practices for defining functions

1.  **Write clear and detailed function names, parameter descriptions, and instructions.**

    - **Explicitly describe the purpose of the function and each parameter** (and its format), and what the output represents.
    - **Use the system prompt to describe when (and when not) to use each function.** Generally, tell the model _exactly_ what to do.
    - **Include examples and edge cases**, especially to rectify any recurring failures. (**Note:** Adding examples may hurt performance for [reasoning models](/docs/guides/reasoning).)

2.  **Apply software engineering best practices.**

    - **Make the functions obvious and intuitive**. ([principle of least surprise](https://en.wikipedia.org/wiki/Principle_of_least_astonishment))
    - **Use enums** and object structure to make invalid states unrepresentable. (e.g. `toggle_light(on: bool, off: bool)` allows for invalid calls)
    - **Pass the intern test.** Can an intern/human correctly use the function given nothing but what you gave the model? (If not, what questions do they ask you? Add the answers to the prompt.)

3.  **Offload the burden from the model and use code where possible.**

    - **Don't make the model fill arguments you already know.** For example, if you already have an `order_id` based on a previous menu, don't have an `order_id` param – instead, have no params `submit_refund()` and pass the `order_id` with code.
    - **Combine functions that are always called in sequence.** For example, if you always call `mark_location()` after `query_location()`, just move the marking logic into the query function call.

4.  **Keep the number of functions small for higher accuracy.**

    - **Evaluate your performance** with different numbers of functions.
    - **Aim for fewer than 20 functions** at any one time, though this is just a soft suggestion.

5.  **Leverage OpenAI resources.**

    - **Generate and iterate on function schemas** in the [Playground](/playground).
    - **Consider [fine-tuning](https://platform.openai.com/docs/guides/fine-tuning) to increase function calling accuracy** for large numbers of functions or difficult tasks. ([cookbook](https://cookbook.openai.com/examples/fine_tuning_for_function_calling))

### Token Usage

Under the hood, functions are injected into the system message in a syntax the model has been trained on. This means functions count against the model's context limit and are billed as input tokens. If you run into token limits, we suggest limiting the number of functions or the length of the descriptions you provide for function parameters.

It is also possible to use [fine-tuning](/docs/guides/fine-tuning#fine-tuning-examples) to reduce the number of tokens used if you have many functions defined in your tools specification.

### Handling function calls

When the model calls a function, you must execute it and return the result. Since model responses can include zero, one, or multiple calls, it is best practice to assume there are several.

The response `output` array contains an entry with the `type` having a value of `function_call`. Each entry with a `call_id` (used later to submit the function result), `name`, and JSON-encoded `arguments`.

Sample response with multiple function calls

```json
[
  {
    "id": "fc_12345xyz",
    "call_id": "call_12345xyz",
    "type": "function_call",
    "name": "get_weather",
    "arguments": "{\"location\":\"Paris, France\"}"
  },
  {
    "id": "fc_67890abc",
    "call_id": "call_67890abc",
    "type": "function_call",
    "name": "get_weather",
    "arguments": "{\"location\":\"Bogotá, Colombia\"}"
  },
  {
    "id": "fc_99999def",
    "call_id": "call_99999def",
    "type": "function_call",
    "name": "send_email",
    "arguments": "{\"to\":\"bob@email.com\",\"body\":\"Hi bob\"}"
  }
]
```

Execute function calls and append results

```javascript
for (const toolCall of response.output) {
  if (toolCall.type !== "function_call") {
    continue;
  }

  const name = toolCall.name;
  const args = JSON.parse(toolCall.arguments);

  const result = callFunction(name, args);
  input.push({
    type: "function_call_output",
    call_id: toolCall.call_id,
    output: result.toString(),
  });
}
```

In the example above, we have a hypothetical `call_function` to route each call. Here’s a possible implementation:

Execute function calls and append results

```javascript
const callFunction = async (name, args) => {
  if (name === "get_weather") {
    return getWeather(args.latitude, args.longitude);
  }
  if (name === "send_email") {
    return sendEmail(args.to, args.body);
  }
};
```

### Formatting results

A result must be a string, but the format is up to you (JSON, error codes, plain text, etc.). The model will interpret that string as needed.

If your function has no return value (e.g. `send_email`), simply return a string to indicate success or failure. (e.g. `"success"`)

### Incorporating results into response

After appending the results to your `input`, you can send them back to the model to get a final response.

Send results back to model

```javascript
const response = await openai.responses.create({
  model: "gpt-4.1",
  input,
  tools,
});
```

Final response

```json
"It's about 15°C in Paris, 18°C in Bogotá, and I've sent that email to Bob."
```

### Additional configurations

### Tool choice

By default the model will determine when and how many tools to use. You can force specific behavior with the `tool_choice` parameter.

1.  **Auto:** (_Default_) Call zero, one, or multiple functions. `tool_choice: "auto"`
2.  **Required:** Call one or more functions. `tool_choice: "required"`

3.  **Forced Function:** Call exactly one specific function. `tool_choice: {"type": "function", "name": "get_weather"}`

![Function Calling Diagram Steps](https://cdn.openai.com/API/docs/images/function-calling-diagram-tool-choice.png)

You can also set `tool_choice` to `"none"` to imitate the behavior of passing no functions.

### Parallel function calling

The model may choose to call multiple functions in a single turn. You can prevent this by setting `parallel_tool_calls` to `false`, which ensures exactly zero or one tool is called.

**Note:** Currently, if you are using a fine tuned model and the model calls multiple functions in one turn then [strict mode](#strict-mode) will be disabled for those calls.

**Note for `gpt-4.1-nano-2025-04-14`:** This snapshot of `gpt-4.1-nano` can sometimes include multiple tools calls for the same tool if parallel tool calls are enabled. It is recommended to disable this feature when using this nano snapshot.

### Strict mode

Setting `strict` to `true` will ensure function calls reliably adhere to the function schema, instead of being best effort. We recommend always enabling strict mode.

Under the hood, strict mode works by leveraging our [structured outputs](/docs/guides/structured-outputs) feature and therefore introduces a couple requirements:

1.  `additionalProperties` must be set to `false` for each object in the `parameters`.
2.  All fields in `properties` must be marked as `required`.

You can denote optional fields by adding `null` as a `type` option (see example below).

Strict mode enabled

```json
{
  "type": "function",
  "name": "get_weather",
  "description": "Retrieves current weather for the given location.",
  "strict": true,
  "parameters": {
    "type": "object",
    "properties": {
      "location": {
        "type": "string",
        "description": "City and country e.g. Bogotá, Colombia"
      },
      "units": {
        "type": ["string", "null"],
        "enum": ["celsius", "fahrenheit"],
        "description": "Units the temperature will be returned in."
      }
    },
    "required": ["location", "units"],
    "additionalProperties": false
  }
}
```

Strict mode disabled

```json
{
  "type": "function",
  "name": "get_weather",
  "description": "Retrieves current weather for the given location.",
  "parameters": {
    "type": "object",
    "properties": {
      "location": {
        "type": "string",
        "description": "City and country e.g. Bogotá, Colombia"
      },
      "units": {
        "type": "string",
        "enum": ["celsius", "fahrenheit"],
        "description": "Units the temperature will be returned in."
      }
    },
    "required": ["location"]
  }
}
```

All schemas generated in the [playground](/playground) have strict mode enabled.

While we recommend you enable strict mode, it has a few limitations:

1.  Some features of JSON schema are not supported. (See [supported schemas](/docs/guides/structured-outputs?context=with_parse#supported-schemas).)

Specifically for fine tuned models:

1.  Schemas undergo additional processing on the first request (and are then cached). If your schemas vary from request to request, this may result in higher latencies.
2.  Schemas are cached for performance, and are not eligible for [zero data retention](/docs/models#how-we-use-your-data).

### Streaming

Streaming can be used to surface progress by showing which function is called as the model fills its arguments, and even displaying the arguments in real time.

Streaming function calls is very similar to streaming regular responses: you set `stream` to `true` and get different `event` objects.

Streaming function calls

```javascript
import { OpenAI } from "openai";

const openai = new OpenAI();

const tools = [
  {
    type: "function",
    name: "get_weather",
    description: "Get current temperature for provided coordinates in celsius.",
    parameters: {
      type: "object",
      properties: {
        latitude: { type: "number" },
        longitude: { type: "number" },
      },
      required: ["latitude", "longitude"],
      additionalProperties: false,
    },
    strict: true,
  },
];

const stream = await openai.responses.create({
  model: "gpt-4.1",
  input: [{ role: "user", content: "What's the weather like in Paris today?" }],
  tools,
  stream: true,
  store: true,
});

for await (const event of stream) {
  console.log(event);
}
```

Output events

```json
{"type":"response.output_item.added","response_id":"resp_1234xyz","output_index":0,"item":{"type":"function_call","id":"fc_1234xyz","call_id":"call_1234xyz","name":"get_weather","arguments":""}}
{"type":"response.function_call_arguments.delta","response_id":"resp_1234xyz","item_id":"fc_1234xyz","output_index":0,"delta":"{\""}
{"type":"response.function_call_arguments.delta","response_id":"resp_1234xyz","item_id":"fc_1234xyz","output_index":0,"delta":"location"}
{"type":"response.function_call_arguments.delta","response_id":"resp_1234xyz","item_id":"fc_1234xyz","output_index":0,"delta":"\":\""}
{"type":"response.function_call_arguments.delta","response_id":"resp_1234xyz","item_id":"fc_1234xyz","output_index":0,"delta":"Paris"}
{"type":"response.function_call_arguments.delta","response_id":"resp_1234xyz","item_id":"fc_1234xyz","output_index":0,"delta":","}
{"type":"response.function_call_arguments.delta","response_id":"resp_1234xyz","item_id":"fc_1234xyz","output_index":0,"delta":" France"}
{"type":"response.function_call_arguments.delta","response_id":"resp_1234xyz","item_id":"fc_1234xyz","output_index":0,"delta":"\"}"}
{"type":"response.function_call_arguments.done","response_id":"resp_1234xyz","item_id":"fc_1234xyz","output_index":0,"arguments":"{\"location\":\"Paris, France\"}"}
{"type":"response.output_item.done","response_id":"resp_1234xyz","output_index":0,"item":{"type":"function_call","id":"fc_1234xyz","call_id":"call_2345abc","name":"get_weather","arguments":"{\"location\":\"Paris, France\"}"}}
```

Instead of aggregating chunks into a single `content` string, however, you're aggregating chunks into an encoded `arguments` JSON object.

When the model calls one or more functions an event of type `response.output_item.added` will be emitted for each function call that contains the following fields:

| Field        | Description                                                                                                   |
| ------------ | ------------------------------------------------------------------------------------------------------------- |
| response_id  | The id of the response that the function call belongs to                                                      |
| output_index | The index of the output item in the response. This respresents the individual function calls in the response. |
| item         | The in-progress function call item that includes a name, arguments and id field                               |

Afterwards you will receive a series of events of type `response.function_call_arguments.delta` which will contain the `delta` of the `arguments` field. These events contain the following fields:

| Field        | Description                                                                                                   |
| ------------ | ------------------------------------------------------------------------------------------------------------- |
| response_id  | The id of the response that the function call belongs to                                                      |
| item_id      | The id of the function call item that the delta belongs to                                                    |
| output_index | The index of the output item in the response. This respresents the individual function calls in the response. |
| delta        | The delta of the arguments field.                                                                             |

Below is a code snippet demonstrating how to aggregate the `delta`s into a final `tool_call` object.

Accumulating tool_call deltas

```javascript
const finalToolCalls = {};

for await (const event of stream) {
  if (event.type === "response.output_item.added") {
    finalToolCalls[event.output_index] = event.item;
  } else if (event.type === "response.function_call_arguments.delta") {
    const index = event.output_index;

    if (finalToolCalls[index]) {
      finalToolCalls[index].arguments += event.delta;
    }
  }
}
```

Accumulated final_tool_calls\[0\]

```json
{
  "type": "function_call",
  "id": "fc_1234xyz",
  "call_id": "call_2345abc",
  "name": "get_weather",
  "arguments": "{\"location\":\"Paris, France\"}"
}
```

When the model has finished calling the functions an event of type `response.function_call_arguments.done` will be emitted. This event contains the entire function call including the following fields:

| Field        | Description                                                                                                   |
| ------------ | ------------------------------------------------------------------------------------------------------------- |
| response_id  | The id of the response that the function call belongs to                                                      |
| output_index | The index of the output item in the response. This respresents the individual function calls in the response. |
| item         | The function call item that includes a name, arguments and id field.                                          |

### Web Search

Allow models to search the web for the latest information before generating a response.

Web search allows models to access up-to-date information from the internet and provide answers with sourced citations. To enable this, use the web search tool in the Responses API or, in some cases, Chat Completions.

There are three main types of web search available with OpenAI models:

1.  Non‑reasoning web search: The non-reasoning model sends the user’s query to the web search tool, which returns the response based on top results. There’s no internal planning and the model simply passes along the search tool’s responses. This method is fast and ideal for quick lookups.
2.  Agentic search with reasoning models is an approach where the model actively manages the search process. It can perform web searches as part of its chain of thought, analyze results, and decide whether to keep searching. This flexibility makes agentic search well suited to complex workflows, but it also means searches take longer than quick lookups. For example, you can adjust GPT-5’s reasoning level to change both the depth and latency of the search.
3.  Deep research is a specialized, agent-driven method for in-depth, extended investigations by reasoning models. The model conducts web searches as part of its chain of thought, often tapping into hundreds of sources. Deep research can run for several minutes and is best used with background mode. These tasks typically use models like `o3-deep-research`, `o4-mini-deep-research`, or `gpt-5` with reasoning level set to `high`.

Using the [Responses API](/docs/api-reference/responses), you can enable web search by configuring it in the `tools` array in an API request to generate content. Like any other tool, the model can choose to search the web or not based on the content of the input prompt.

Web search tool example

```javascript
import OpenAI from "openai";
const client = new OpenAI();

const response = await client.responses.create({
  model: "gpt-5",
  tools: [{ type: "web_search" }],
  input: "What was a positive news story from today?",
});

console.log(response.output_text);
```

```python
from openai import OpenAI
client = OpenAI()

response = client.responses.create(
    model="gpt-5",
    tools=[{"type": "web_search"}],
    input="What was a positive news story from today?"
)

print(response.output_text)
```

```bash
curl "https://api.openai.com/v1/responses" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    -d '{
        "model": "gpt-5",
        "tools": [{"type": "web_search"}],
        "input": "what was a positive news story from today?"
    }'
```

```csharp
using OpenAI.Responses;

string key = Environment.GetEnvironmentVariable("OPENAI_API_KEY")!;
OpenAIResponseClient client = new(model: "gpt-5", apiKey: key);

ResponseCreationOptions options = new();
options.Tools.Add(ResponseTool.CreateWebSearchTool());

OpenAIResponse response = (OpenAIResponse)client.CreateResponse([
    ResponseItem.CreateUserMessageItem([
        ResponseContentPart.CreateInputTextPart("What was a positive news story from today?"),
    ]),
], options);

Console.WriteLine(response.GetOutputText());
```

## Output and citations

Model responses that use the web search tool will include two parts:

- A `web_search_call` output item with the ID of the search call, along with the action taken in `web_search_call.action`. The action is one of:
  - `search`, which represents a web search. It will usually (but not always) includes the search `query` and `domains` which were searched. Search actions incur a tool call cost (see [pricing](/docs/pricing#built-in-tools)).
  - `open_page`, which represents a page being opened. Supported in reasoning models.
  - `find_in_page`, which represents searching within a page. Supported in reasoning models.
- A `message` output item containing:
  - The text result in `message.content[0].text`
  - Annotations `message.content[0].annotations` for the cited URLs

By default, the model's response will include inline citations for URLs found in the web search results. In addition to this, the `url_citation` annotation object will contain the URL, title and location of the cited source.

When displaying web results or information contained in web results to end users, inline citations must be made clearly visible and clickable in your user interface.

```json
[
  {
    "type": "web_search_call",
    "id": "ws_67c9fa0502748190b7dd390736892e100be649c1a5ff9609",
    "status": "completed"
  },
  {
    "id": "msg_67c9fa077e288190af08fdffda2e34f20be649c1a5ff9609",
    "type": "message",
    "status": "completed",
    "role": "assistant",
    "content": [
      {
        "type": "output_text",
        "text": "On March 6, 2025, several news...",
        "annotations": [
          {
            "type": "url_citation",
            "start_index": 2606,
            "end_index": 2758,
            "url": "https://...",
            "title": "Title..."
          }
        ]
      }
    ]
  }
]
```

## Domain filtering

Domain filtering in web search lets you limit results to a specific set of domains. With the `filters` parameter you can set an allow-list of up to 20 URLs. When formatting URLs, omit the HTTP or HTTPS prefix. For example, use [`openai.com`](http://openai.com) instead of [`https://openai.com/`](https://openai.com/). This approach also includes subdomains in the search. Note that domain filtering is only available in the Responses API with the `web_search` tool.

## Sources

To view all URLs retrieved during a web search, use the `sources` field. Unlike inline citations, which show only the most relevant references, sources returns the complete list of URLs the model consulted when forming its response. The number of sources is often greater than the number of citations. Real-time third-party feeds are also surfaced here and are labeled as `oai-sports`, `oai-weather`, or `oai-finance`. The sources field is available with both the `web_search` and `web_search_preview` tools.

List sources

```bash
curl "https://api.openai.com/v1/responses" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $OPENAI_API_KEY" \
-d '{
  "model": "gpt-5",
  "reasoning": { "effort": "low" },
  "tools": [
    {
      "type": "web_search",
      "filters": {
        "allowed_domains": [
          "pubmed.ncbi.nlm.nih.gov",
          "clinicaltrials.gov",
          "www.who.int",
          "www.cdc.gov",
          "www.fda.gov"
        ]
      }
    }
  ],
  "tool_choice": "auto",
  "include": ["web_search_call.action.sources"],
  "input": "Please perform a web search on how semaglutide is used in the treatment of diabetes."
}'
```

```javascript
import OpenAI from "openai";
const client = new OpenAI();

const response = await client.responses.create({
  model: "gpt-5",
  reasoning: { effort: "low" },
  tools: [
    {
      type: "web_search",
      filters: {
        allowed_domains: [
          "pubmed.ncbi.nlm.nih.gov",
          "clinicaltrials.gov",
          "www.who.int",
          "www.cdc.gov",
          "www.fda.gov",
        ],
      },
    },
  ],
  tool_choice: "auto",
  include: ["web_search_call.action.sources"],
  input:
    "Please perform a web search on how semaglutide is used in the treatment of diabetes.",
});

console.log(response.output_text);
```

```python
from openai import OpenAI
client = OpenAI()

response = client.responses.create(
  model="gpt-5",
  reasoning={"effort": "low"},
  tools=[
      {
          "type": "web_search",
          "filters": {
              "allowed_domains": [
                  "pubmed.ncbi.nlm.nih.gov",
                  "clinicaltrials.gov",
                  "www.who.int",
                  "www.cdc.gov",
                  "www.fda.gov",
              ]
          },
      }
  ],
  tool_choice="auto",
  include=["web_search_call.action.sources"],
  input="Please perform a web search on how semaglutide is used in the treatment of diabetes.",
)

print(response.output_text)
```

## User location

To refine search results based on geography, you can specify an approximate user location using country, city, region, and/or timezone.

- The `city` and `region` fields are free text strings, like `Minneapolis` and `Minnesota` respectively.
- The `country` field is a two-letter [ISO country code](https://en.wikipedia.org/wiki/ISO_3166-1), like `US`.
- The `timezone` field is an [IANA timezone](https://timeapi.io/documentation/iana-timezones) like `America/Chicago`.

Note that user location is not supported for deep research models using web search.

Customizing user location

```python
from openai import OpenAI
client = OpenAI()

response = client.responses.create(
    model="o4-mini",
    tools=[{
        "type": "web_search",
        "user_location": {
            "type": "approximate",
            "country": "GB",
            "city": "London",
            "region": "London",
        }
    }],
    input="What are the best restaurants around Granary Square?",
)

print(response.output_text)
```

```javascript
import OpenAI from "openai";
const openai = new OpenAI();

const response = await openai.responses.create({
  model: "o4-mini",
  tools: [
    {
      type: "web_search",
      user_location: {
        type: "approximate",
        country: "GB",
        city: "London",
        region: "London",
      },
    },
  ],
  input: "What are the best restaurants around Granary Square?",
});
console.log(response.output_text);
```

```bash
curl "https://api.openai.com/v1/responses" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    -d '{
        "model": "o4-mini",
        "tools": [{
            "type": "web_search",
            "user_location": {
                "type": "approximate",
                "country": "GB",
                "city": "London",
                "region": "London"
            }
        }],
        "input": "What are the best restaurants around Granary Square?"
    }'
```

## API compatibility

Web search is available in the Responses API as the generally available version of the tool, `web_search`, as well as the earlier tool version, `web_search_preview`. To use web search in the Chat Completions API, use the specialized web search models `gpt-4o-search-preview` and `gpt-4o-mini-search-preview`.

## Limitations

- Web search is currently not supported in [`gpt-5`](/docs/models/gpt-5) with `minimal` reasoning, and [`gpt-4.1-nano`](/docs/models/gpt-4.1-nano).
- When used as a tool in the [Responses API](/docs/api-reference/responses), web search has the same tiered rate limits as the models above.
- Web search is limited to a context window size of 128000 (even with [`gpt-4.1`](/docs/models/gpt-4.1) and [`gpt-4.1-mini`](/docs/models/gpt-4.1-mini) models).

## Usage notes

||
|ResponsesChat CompletionsAssistants|Same as tiered rate limits for underlying model used with the tool.|PricingZDR and data residency|

Was this page useful?
