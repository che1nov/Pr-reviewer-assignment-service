import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const successRate = new Rate('success');
const responseTime = new Trend('response_time');

export const options = {
  scenarios: {
    constant_load: {
      executor: 'constant-arrival-rate',
      rate: 5, 
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 10,
      maxVUs: 50,
    },
  },
  
  thresholds: {
    'http_req_duration': ['p(95)<300'], 
    'http_req_duration{endpoint:health}': ['p(95)<100'], 
    'http_req_duration{endpoint:create_team}': ['p(95)<300'],
    'http_req_duration{endpoint:create_pr}': ['p(95)<300'],
    'http_req_duration{endpoint:stats}': ['p(95)<300'],
    'http_req_duration{endpoint:deactivate_team}': ['p(95)<100'],
    'http_req_failed': ['rate<0.001'],
    'errors': ['rate<0.001'],
    'success': ['rate>0.999'],
  },
};

const BASE_URL = 'http://localhost:8080';
const ADMIN_TOKEN = 'admin-secret';
const USER_TOKEN = 'user-secret';

let teamCounter = 0;
let prCounter = 0;

export default function () {
  const batch = http.batch([
    ['GET', `${BASE_URL}/health`, null, {
      tags: { endpoint: 'health' },
    }],
    
    ['GET', `${BASE_URL}/stats`, null, {
      headers: { 'Authorization': `Bearer ${USER_TOKEN}` },
      tags: { endpoint: 'stats' },
    }],
  ]);

  batch.forEach((response, index) => {
    const success = check(response, {
      'status is 200': (r) => r.status === 200,
      'response time < 300ms': (r) => r.timings.duration < 300,
    });
    
    errorRate.add(!success);
    successRate.add(success);
    responseTime.add(response.timings.duration);
  });

  if (Math.random() < 0.3) { 
    teamCounter++;
    const teamName = `team-${Date.now()}-${teamCounter}`;
    
    const teamPayload = JSON.stringify({
      team_name: teamName,
      members: [
        { user_id: `user1-${teamCounter}`, username: 'Alice', is_active: true },
        { user_id: `user2-${teamCounter}`, username: 'Bob', is_active: true },
        { user_id: `user3-${teamCounter}`, username: 'Charlie', is_active: true },
      ],
    });

    const createTeamRes = http.post(`${BASE_URL}/team/add`, teamPayload, {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${ADMIN_TOKEN}`,
      },
      tags: { endpoint: 'create_team' },
    });

    const teamSuccess = check(createTeamRes, {
      'team created': (r) => r.status === 201,
      'team response time < 300ms': (r) => r.timings.duration < 300,
    });
    
    errorRate.add(!teamSuccess);
    successRate.add(teamSuccess);
    responseTime.add(createTeamRes.timings.duration);

    if (createTeamRes.status === 201) {
      sleep(0.1);
      
      prCounter++;
      const prPayload = JSON.stringify({
        pull_request_id: `pr-${Date.now()}-${prCounter}`,
        pull_request_name: `Feature ${prCounter}`,
        author_id: `user1-${teamCounter}`,
      });

      const createPRRes = http.post(`${BASE_URL}/pullRequest/create`, prPayload, {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${ADMIN_TOKEN}`,
        },
        tags: { endpoint: 'create_pr' },
      });

      const prSuccess = check(createPRRes, {
        'PR created': (r) => r.status === 201,
        'PR response time < 300ms': (r) => r.timings.duration < 300,
      });
      
      errorRate.add(!prSuccess);
      successRate.add(prSuccess);
      responseTime.add(createPRRes.timings.duration);

      if (Math.random() < 0.2) {
        sleep(0.1);
        
        const deactivatePayload = JSON.stringify({
          team_name: teamName,
        });

        const deactivateRes = http.post(`${BASE_URL}/team/deactivateUsers`, deactivatePayload, {
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${ADMIN_TOKEN}`,
          },
          tags: { endpoint: 'deactivate_team' },
        });

        const deactivateSuccess = check(deactivateRes, {
          'deactivate successful': (r) => r.status === 200,
          'deactivate time < 100ms': (r) => r.timings.duration < 100,
        });
        
        errorRate.add(!deactivateSuccess);
        successRate.add(deactivateSuccess);
        responseTime.add(deactivateRes.timings.duration);
      }
    }
  }

  sleep(0.1);
}

export function handleSummary(data) {
  return {
    'load_test_results.json': JSON.stringify(data, null, 2),
    stdout: textSummary(data, { indent: ' ', enableColors: true }),
  };
}

function textSummary(data, options) {
  const indent = options.indent || '';
  const colors = options.enableColors;
  
  summary += indent + '  НАГРУЗОЧНОЕ ТЕСТИРОВАНИЕ - РЕЗУЛЬТАТЫ\n';
  
  const metrics = data.metrics;
  
  // Общая статистика
  summary += indent + 'Общая статистика:\n';
  summary += indent + `  • Всего запросов: ${metrics.http_reqs.values.count}\n`;
  summary += indent + `  • Успешных: ${(metrics.success.values.rate * 100).toFixed(2)}%\n`;
  summary += indent + `  • Ошибок: ${(metrics.errors.values.rate * 100).toFixed(3)}%\n`;
  summary += indent + `  • RPS: ${metrics.http_reqs.values.rate.toFixed(2)}\n\n`;
  
  // Время ответа
  summary += indent + 'Время ответа (ms):\n';
  summary += indent + `  • Среднее: ${metrics.http_req_duration.values.avg.toFixed(2)}ms\n`;
  summary += indent + `  • Медиана: ${metrics.http_req_duration.values.med.toFixed(2)}ms\n`;
  summary += indent + `  • P95: ${metrics.http_req_duration.values['p(95)'].toFixed(2)}ms\n`;
  summary += indent + `  • P99: ${metrics.http_req_duration.values['p(99)'].toFixed(2)}ms\n`;
  summary += indent + `  • Максимум: ${metrics.http_req_duration.values.max.toFixed(2)}ms\n\n`;
  
  // Проверка SLI
  summary += indent + 'Проверка SLI:\n';
  const sliRPS = metrics.http_reqs.values.rate >= 5;
  const sliLatency = metrics.http_req_duration.values['p(95)'] < 300;
  const sliSuccess = metrics.http_req_failed.values.rate < 0.001;
  
  summary += indent + `  • RPS ≥ 5: ${sliRPS ? '✓' : '✗'} (${metrics.http_reqs.values.rate.toFixed(2)})\n`;
  summary += indent + `  • P95 < 300ms: ${sliLatency ? '✓' : '✗'} (${metrics.http_req_duration.values['p(95)'].toFixed(2)}ms)\n`;
  summary += indent + `  • Успешность ≥ 99.9%: ${sliSuccess ? '✓' : '✗'} (${((1 - metrics.http_req_failed.values.rate) * 100).toFixed(2)}%)\n\n`;
  
  const allPassed = sliRPS && sliLatency && sliSuccess;
  if (allPassed) {
    summary += indent + 'ВСЕ SLI ВЫПОЛНЕНЫ!\n';
  } else {
    summary += indent + 'НЕКОТОРЫЕ SLI НЕ ВЫПОЛНЕНЫ\n';
  }
  summary += indent + '═══════════════════════════════════════════\n\n';
  
  return summary;
}

