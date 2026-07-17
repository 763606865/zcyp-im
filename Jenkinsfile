pipeline {
  agent any

  options {
    timestamps()
    disableConcurrentBuilds()
  }

  environment {
    GO111MODULE = 'on'
    CGO_ENABLED = '0'
    APP_NAME = 'zcyp-im'
    ARTIFACTS_DIR = 'artifacts'
    PACKAGE_SCRIPT = 'scripts/jenkins/package_release.sh'
    DEPLOY_SCRIPT = 'scripts/jenkins/deploy_release.sh'
    RUN_TESTS = "${env.RUN_TESTS ?: 'true'}"
    RUN_MIGRATIONS = "${env.RUN_MIGRATIONS ?: 'true'}"
    DEPLOY_ROOT = "${env.DEPLOY_ROOT ?: '/home/data/go/zcyp-im'}"
    SERVICE_API = "${env.SERVICE_API ?: 'zcyp-im-api'}"
    SERVICE_WS = "${env.SERVICE_WS ?: 'zcyp-im-ws'}"
    SHARED_ENV_FILE = "${env.SHARED_ENV_FILE ?: '/home/data/go/zcyp-im/shared/.env'}"
  }

  stages {
    stage('Checkout') {
      steps {
        checkout scm
      }
    }

    stage('Package') {
      steps {
        sh 'chmod +x ${PACKAGE_SCRIPT} ${DEPLOY_SCRIPT}'
        sh '${PACKAGE_SCRIPT}'
      }
    }

    stage('Deploy') {
      when {
        expression { env.SKIP_DEPLOY != 'true' }
      }
      steps {
        sh '${DEPLOY_SCRIPT} "artifacts/${APP_NAME}-${BUILD_NUMBER}.tar.gz"'
      }
    }
  }

  post {
    always {
      archiveArtifacts artifacts: 'artifacts/*.tar.gz', fingerprint: true, onlyIfSuccessful: false
    }
  }
}
