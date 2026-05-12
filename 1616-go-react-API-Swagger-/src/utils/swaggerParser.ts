import { SkippedApi, ParameterInfo, RequestBodyInfo, ResponseInfo } from '../types';
import { v4 as uuidv4 } from 'uuid';

const VALID_METHODS = ['get', 'post', 'put', 'delete', 'patch', 'head', 'options', 'trace'];

interface ParsedApi {
  id: string;
  path: string;
  method: string;
  summary?: string;
  description?: string;
  tags: string[];
  parameters: ParameterInfo[];
  requestBody?: RequestBodyInfo;
  responses: ResponseInfo[];
}

interface ParseResult {
  apis: ParsedApi[];
  skipped_apis: SkippedApi[];
  baseUrl?: string;
  serviceName?: string;
  serviceDescription?: string;
}

const parseSwagger = (swaggerJson: any): ParseResult => {
  const skippedApis: SkippedApi[] = [];
  const apis: ParsedApi[] = [];

  if (!swaggerJson || typeof swaggerJson !== 'object') {
    throw new Error('Invalid Swagger/OpenAPI document');
  }

  const isOpenAPI3 = 'openapi' in swaggerJson && swaggerJson.openapi?.startsWith('3');

  const result: ParseResult = {
    apis: [],
    skipped_apis: [],
  };

  if (swaggerJson.info) {
    result.serviceName = swaggerJson.info.title;
    result.serviceDescription = swaggerJson.info.description;
  }

  if (isOpenAPI3 && swaggerJson.servers?.length > 0) {
    result.baseUrl = swaggerJson.servers[0].url;
  }

  const paths = swaggerJson.paths || {};

  for (const [path, pathItem] of Object.entries(paths)) {
    if (!pathItem || typeof pathItem !== 'object') {
      continue;
    }

    for (const method of VALID_METHODS) {
      const operation = (pathItem as any)[method];
      if (!operation || typeof operation !== 'object') {
        continue;
      }

      try {
        const api = parseOperation(path, method.toUpperCase(), operation, pathItem as any, isOpenAPI3);
        apis.push(api);
      } catch (error: any) {
        skippedApis.push({
          path,
          method: method.toUpperCase(),
          reason: error.message || 'Unknown error',
        });
      }
    }
  }

  result.apis = apis;
  result.skipped_apis = skippedApis;

  return result;
};

const parseOperation = (
  path: string,
  method: string,
  operation: any,
  pathItem: any,
  isOpenAPI3: boolean
): ParsedApi => {
  const parameters: ParameterInfo[] = [];
  const tags = operation.tags || [];

  const allParams = [
    ...(pathItem.parameters || []),
    ...(operation.parameters || []),
  ];

  for (const param of allParams) {
    if (!param.name || !param.in) {
      throw new Error(`Invalid parameter definition: missing name or location`);
    }
    parameters.push({
      name: param.name,
      in: param.in,
      required: param.required || false,
      type: param.type,
      schema: param.schema,
      description: param.description,
    });
  }

  let requestBody: RequestBodyInfo | undefined;
  if (operation.requestBody) {
    requestBody = {
      description: operation.requestBody.description,
      content: operation.requestBody.content,
      required: operation.requestBody.required,
    };
  }

  const responses: ResponseInfo[] = [];
  const responseObj = operation.responses || {};
  for (const [statusCode, response] of Object.entries(responseObj)) {
    responses.push({
      statusCode,
      description: (response as any).description,
      content: (response as any).content,
    });
  }

  return {
    id: uuidv4(),
    path,
    method,
    summary: operation.summary,
    description: operation.description,
    tags: Array.isArray(tags) ? tags : [],
    parameters,
    requestBody,
    responses,
  };
};

const validateJson = (jsonString: string): any => {
  try {
    return JSON.parse(jsonString);
  } catch (e: any) {
    const match = e.message.match(/position (\d+)/);
    if (match) {
      const position = parseInt(match[1], 10);
      const before = jsonString.slice(Math.max(0, position - 20), position);
      const after = jsonString.slice(position, position + 20);
      throw new Error(`JSON parse error at position ${position}: "...${before}|${after}..."`);
    }
    throw new Error(`JSON parse error: ${e.message}`);
  }
};

export { parseSwagger, validateJson };
export type { ParsedApi, ParseResult };
