<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ApiError, apiFetch } from '../api'
import type { CreateProjectRequest, CurrentUser, Project } from '../types'

const props = defineProps<{ user: CurrentUser }>()
const router = useRouter()
const steps = ['프로젝트 범위', '기본 정보', '저장소 연결', '검토 및 생성']
const currentStep = ref(0)
const submitting = ref(false)
const confirmed = ref(false)
const submitError = ref('')
const createdProject = ref<Project>()
const errors = reactive<Record<string, string>>({})
const touched = reactive<Record<string, boolean>>({})

const form = reactive<CreateProjectRequest>({
  scope: 'personal',
  organizationId: null,
  name: '',
  slug: '',
  description: '',
  repositoryAlias: '',
  repositorySourceType: 'local_path',
  repositoryPath: '',
  repositoryUrl: '',
  credentialRef: '',
  defaultBranch: 'main',
  excludePatterns: ['node_modules/**', 'dist/**', 'vendor/**'],
})
const excludeText = ref(form.excludePatterns.join('\n'))
const slugEdited = ref(false)
const organizations = computed(() => props.user.organizations.filter((organization) => organization.status === 'active'))
const selectedOrganization = computed(() => organizations.value.find((organization) => organization.id === form.organizationId))
const scopeLabel = computed(() => form.scope === 'personal' ? '개인 프로젝트' : selectedOrganization.value?.name || '조직 프로젝트')
const repositoryLocation = computed(() => form.repositorySourceType === 'local_path' ? form.repositoryPath : form.repositoryUrl)

function slugify(value: string): string {
  return value.toLowerCase().trim().replace(/[^a-z0-9가-힣]+/g, '-').replace(/[가-힣]/g, '').replace(/^-+|-+$/g, '').slice(0, 63)
}

watch(() => form.name, (name) => {
  if (!slugEdited.value) form.slug = slugify(name)
  if (!form.repositoryAlias) form.repositoryAlias = name.trim().slice(0, 120)
})

watch(excludeText, (value) => {
  form.excludePatterns = value.split('\n').map((item) => item.trim()).filter(Boolean)
})

watch(() => form.scope, (scope) => {
  if (scope === 'personal') form.organizationId = null
  else if (!form.organizationId) form.organizationId = organizations.value[0]?.id || null
})

onMounted(() => {
  const saved = sessionStorage.getItem('cosight.project-draft')
  if (!saved) return
  try {
    const draft = JSON.parse(saved) as Partial<CreateProjectRequest>
    Object.assign(form, draft)
    excludeText.value = (draft.excludePatterns || form.excludePatterns).join('\n')
    slugEdited.value = Boolean(draft.slug)
  } catch { sessionStorage.removeItem('cosight.project-draft') }
})

watch(form, (value) => {
  if (!createdProject.value) sessionStorage.setItem('cosight.project-draft', JSON.stringify(value))
}, { deep: true })

function validateField(field: string): string {
  const value = String(form[field as keyof CreateProjectRequest] ?? '').trim()
  if (field === 'organizationId' && form.scope === 'organization' && !form.organizationId) return '프로젝트를 소유할 조직을 선택해 주세요.'
  if (field === 'name' && !value) return '프로젝트 이름을 입력해 주세요.'
  if (field === 'name' && value.length > 120) return '프로젝트 이름은 120자 이하로 입력해 주세요.'
  if (field === 'slug' && !/^[a-z0-9][a-z0-9-]{1,62}$/.test(value)) return '2~63자의 영문 소문자, 숫자, 하이픈만 사용할 수 있습니다.'
  if (field === 'description' && value.length > 2000) return '설명은 2,000자 이하로 입력해 주세요.'
  if (field === 'repositoryAlias' && !value) return '저장소 표시 이름을 입력해 주세요.'
  if (field === 'repositoryPath' && form.repositorySourceType === 'local_path') {
    if (!value) return '서버의 저장소 절대 경로를 입력해 주세요.'
    if (!value.startsWith('/')) return '절대 경로로 입력해 주세요. 예: /repositories/cosight'
  }
  if (field === 'repositoryUrl' && form.repositorySourceType === 'git_url') {
    if (!value) return 'Git 저장소 URL을 입력해 주세요.'
    try {
      const url = new URL(value)
      if (!['https:', 'ssh:'].includes(url.protocol)) return 'HTTPS 또는 SSH URL만 사용할 수 있습니다.'
    } catch { return '올바른 Git 저장소 URL을 입력해 주세요.' }
  }
  if (field === 'defaultBranch' && !value) return '기본 브랜치를 입력해 주세요.'
  if (field === 'excludePatterns' && form.excludePatterns.length > 100) return '제외 패턴은 최대 100개까지 입력할 수 있습니다.'
  return ''
}

const fieldsByStep: string[][] = [
  ['organizationId'],
  ['name', 'slug', 'description'],
  ['repositoryAlias', 'repositoryPath', 'repositoryUrl', 'defaultBranch', 'excludePatterns'],
]

function validateStep(step: number): boolean {
  const fields = fieldsByStep[step] || fieldsByStep.flat()
  let valid = true
  for (const field of fields) {
    touched[field] = true
    errors[field] = validateField(field)
    if (errors[field]) valid = false
  }
  return valid
}

function touch(field: string): void {
  touched[field] = true
  errors[field] = validateField(field)
}

function next(): void {
  if (!validateStep(currentStep.value)) return
  currentStep.value = Math.min(currentStep.value + 1, steps.length - 1)
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function back(): void {
  currentStep.value = Math.max(currentStep.value - 1, 0)
  submitError.value = ''
}

function goToStep(step: number): void {
  if (step < currentStep.value) currentStep.value = step
}

function requestBody(): CreateProjectRequest {
  return {
    ...form,
    organizationId: form.scope === 'organization' ? form.organizationId : null,
    description: form.description?.trim() || null,
    repositoryPath: form.repositorySourceType === 'local_path' ? form.repositoryPath?.trim() : undefined,
    repositoryUrl: form.repositorySourceType === 'git_url' ? form.repositoryUrl?.trim() : undefined,
    credentialRef: form.repositorySourceType === 'git_url' ? form.credentialRef?.trim() || null : null,
    defaultBranch: form.defaultBranch?.trim() || null,
  }
}

async function createProject(): Promise<void> {
  if (!confirmed.value) {
    submitError.value = '저장소 접근 권한과 공개 범위 확인에 동의해 주세요.'
    return
  }
  if (!validateStep(-1)) {
    currentStep.value = fieldsByStep.findIndex((fields) => fields.some((field) => errors[field]))
    return
  }
  submitting.value = true
  submitError.value = ''
  try {
    createdProject.value = await apiFetch<Project>('/api/v1/projects', {
      method: 'POST',
      body: JSON.stringify(requestBody()),
    })
    sessionStorage.removeItem('cosight.project-draft')
  } catch (error) {
    if (error instanceof ApiError) {
      const serverFieldErrors = { ...error.fieldErrors }
      if (!Object.keys(serverFieldErrors).length && error.status === 409) serverFieldErrors.slug = '같은 범위에 이미 사용 중인 slug입니다.'
      if (error.code.includes('REPOSITORY_PATH')) serverFieldErrors.repositoryPath = error.message
      Object.assign(errors, serverFieldErrors)
      submitError.value = error.message
      const invalidStep = fieldsByStep.findIndex((fields) => fields.some((field) => serverFieldErrors[field]))
      if (invalidStep >= 0) currentStep.value = invalidStep
    } else submitError.value = error instanceof Error ? error.message : '프로젝트를 만들지 못했습니다.'
  } finally { submitting.value = false }
}

function openProject(): void {
  if (createdProject.value) router.push(`/projects/${createdProject.value.id}`)
}
</script>

<template>
  <div class="page-heading create-heading">
    <div><p>프로젝트</p><h1>프로젝트 만들기</h1><span>저장소를 연결하고 첫 코드 분석을 준비합니다.</span></div>
    <RouterLink class="quiet-link" to="/">취소하고 목록으로</RouterLink>
  </div>

  <div v-if="createdProject" class="creation-success" role="status">
    <span class="success-mark">✓</span>
    <p>프로젝트 생성 완료</p>
    <h2>{{ createdProject.name }}</h2>
    <span><b>Project Admin</b> 권한이 자동으로 부여되었습니다.</span>
    <div class="success-summary"><div><small>프로젝트 범위</small><strong>{{ scopeLabel }}</strong></div><div><small>저장소</small><strong>{{ form.repositoryAlias }}</strong></div><div><small>상태</small><strong>인덱싱 대기</strong></div></div>
    <div class="info-callout"><span>i</span><p><b>코드 인덱싱은 아직 시작되지 않았습니다.</b><br>프로젝트 개요에서 저장소 연결 상태를 확인한 뒤 첫 인덱싱을 실행할 수 있습니다.</p></div>
    <button class="primary-button form-button" @click="openProject">프로젝트 개요로 이동 <span>→</span></button>
  </div>

  <div v-else class="create-layout">
    <aside class="step-panel" aria-label="프로젝트 생성 단계">
      <ol>
        <li v-for="(step, index) in steps" :key="step" :class="{ active: index === currentStep, done: index < currentStep }">
          <button :disabled="index > currentStep" @click="goToStep(index)"><span>{{ index < currentStep ? '✓' : index + 1 }}</span><div><small>STEP {{ index + 1 }}</small><b>{{ step }}</b></div></button>
        </li>
      </ol>
      <div class="draft-note"><span>✓</span><p><b>입력 내용 자동 저장</b><small>이 브라우저에서 다시 이어갈 수 있습니다.</small></p></div>
    </aside>

    <section class="form-card">
      <header><span>{{ currentStep + 1 }} / {{ steps.length }}</span><h2>{{ steps[currentStep] }}</h2><p v-if="currentStep === 0">누가 프로젝트를 소유하고 관리할지 선택합니다.</p><p v-else-if="currentStep === 1">프로젝트를 식별할 이름과 URL용 slug를 정합니다.</p><p v-else-if="currentStep === 2">Cosight가 읽기 전용으로 분석할 저장소를 연결합니다.</p><p v-else>입력한 내용을 확인하면 프로젝트가 생성됩니다.</p></header>

      <div v-if="submitError" class="form-alert" role="alert"><span>!</span><div><b>프로젝트를 만들 수 없습니다.</b><p>{{ submitError }}</p></div></div>

      <div v-if="currentStep === 0" class="form-section">
        <fieldset class="choice-fieldset"><legend>프로젝트 범위</legend>
          <label class="choice-card" :class="{ selected: form.scope === 'personal' }"><input v-model="form.scope" type="radio" value="personal"><span class="choice-icon">♙</span><div><b>개인 프로젝트</b><p>나를 소유자로 생성합니다. 이후 다른 사용자를 초대할 수 있습니다.</p><small>모든 사용자 사용 가능</small></div><span class="radio-dot"></span></label>
          <label v-if="organizations.length" class="choice-card" :class="{ selected: form.scope === 'organization' }"><input v-model="form.scope" type="radio" value="organization"><span class="choice-icon">▦</span><div><b>조직 프로젝트</b><p>선택한 조직이 소유하며 조직 구성원만 초대할 수 있습니다.</p><small>조직 소속 사용자만 사용 가능</small></div><span class="radio-dot"></span></label>
        </fieldset>
        <label v-if="form.scope === 'organization'" class="field-label"><span>소유 조직 <em>*</em></span><select v-model="form.organizationId" :class="{ invalid: touched.organizationId && errors.organizationId }" @blur="touch('organizationId')"><option :value="null" disabled>조직 선택</option><option v-for="organization in organizations" :key="organization.id" :value="organization.id">{{ organization.name }}</option></select><small v-if="errors.organizationId" class="field-error">{{ errors.organizationId }}</small></label>
        <div v-if="!organizations.length" class="inline-hint"><span>i</span><p>현재 소속된 조직이 없어 개인 프로젝트로 생성됩니다. 조직 프로젝트는 시스템 관리자가 조직에 사용자를 할당한 뒤 만들 수 있습니다.</p></div>
      </div>

      <div v-else-if="currentStep === 1" class="form-section">
        <label class="field-label"><span>프로젝트 이름 <em>*</em></span><input v-model="form.name" maxlength="120" placeholder="예: 결제 서비스" :class="{ invalid: touched.name && errors.name }" autofocus @blur="touch('name')"><small v-if="errors.name" class="field-error">{{ errors.name }}</small><small v-else>목록과 헤더에 표시되는 이름입니다.</small></label>
        <label class="field-label"><span>프로젝트 slug <em>*</em></span><div class="input-prefix"><span>/projects/</span><input v-model="form.slug" maxlength="63" placeholder="billing-service" :class="{ invalid: touched.slug && errors.slug }" @input="slugEdited = true" @blur="touch('slug')"></div><small v-if="errors.slug" class="field-error">{{ errors.slug }}</small><small v-else>범위 안에서 고유해야 하며 생성 후에도 변경할 수 있습니다.</small></label>
        <label class="field-label"><span>설명 <i>선택</i></span><textarea v-model="form.description" maxlength="2000" rows="4" placeholder="팀이 이 프로젝트에서 분석할 코드와 목적을 설명해 주세요." @blur="touch('description')"></textarea><small :class="{ 'field-error': errors.description }">{{ errors.description || `${form.description?.length || 0} / 2,000` }}</small></label>
      </div>

      <div v-else-if="currentStep === 2" class="form-section">
        <div class="two-columns"><label class="field-label"><span>저장소 표시 이름 <em>*</em></span><input v-model="form.repositoryAlias" maxlength="120" placeholder="billing-main" :class="{ invalid: errors.repositoryAlias }" @blur="touch('repositoryAlias')"><small v-if="errors.repositoryAlias" class="field-error">{{ errors.repositoryAlias }}</small></label><label class="field-label"><span>기본 브랜치 <em>*</em></span><input v-model="form.defaultBranch" placeholder="main" :class="{ invalid: errors.defaultBranch }" @blur="touch('defaultBranch')"><small v-if="errors.defaultBranch" class="field-error">{{ errors.defaultBranch }}</small></label></div>
        <fieldset class="source-tabs"><legend>저장소 연결 방식</legend><label :class="{ selected: form.repositorySourceType === 'local_path' }"><input v-model="form.repositorySourceType" type="radio" value="local_path"><b>서버 경로</b><small>서버에 마운트된 저장소</small></label><label :class="{ selected: form.repositorySourceType === 'git_url' }"><input v-model="form.repositorySourceType" type="radio" value="git_url"><b>Git URL</b><small>원격 저장소 clone</small></label></fieldset>
        <label v-if="form.repositorySourceType === 'local_path'" class="field-label"><span>저장소 절대 경로 <em>*</em></span><input v-model="form.repositoryPath" class="code-input" placeholder="/repositories/team/billing-service" :class="{ invalid: errors.repositoryPath }" @blur="touch('repositoryPath')"><small v-if="errors.repositoryPath" class="field-error">{{ errors.repositoryPath }}</small><small v-else>허용된 저장소 루트 아래의 읽기 전용 경로만 사용할 수 있습니다. symlink는 서버에서 다시 검증합니다.</small></label>
        <template v-else><label class="field-label"><span>Git 저장소 URL <em>*</em></span><input v-model="form.repositoryUrl" class="code-input" placeholder="https://git.example.com/team/billing.git" :class="{ invalid: errors.repositoryUrl }" @blur="touch('repositoryUrl')"><small v-if="errors.repositoryUrl" class="field-error">{{ errors.repositoryUrl }}</small></label><label class="field-label"><span>자격증명 참조 <i>선택</i></span><input v-model="form.credentialRef" class="code-input" placeholder="vault://repositories/billing"><small>비밀값이 아닌 Secret Manager 참조만 입력합니다.</small></label></template>
        <label class="field-label"><span>분석 제외 패턴 <i>선택 · 줄바꿈으로 구분</i></span><textarea v-model="excludeText" class="code-input" rows="5" placeholder="node_modules/**&#10;dist/**"></textarea><small v-if="errors.excludePatterns" class="field-error">{{ errors.excludePatterns }}</small><small v-else>{{ form.excludePatterns.length }}개 패턴 · glob 형식을 사용합니다.</small></label>
      </div>

      <div v-else class="form-section review-section">
        <div class="review-group"><div class="review-title"><span class="review-icon">♙</span><div><small>프로젝트 범위</small><b>{{ scopeLabel }}</b></div><button @click="goToStep(0)">수정</button></div><dl><div><dt>생성자 역할</dt><dd>Project Admin</dd></div><div><dt>공유 범위</dt><dd>{{ form.scope === 'personal' ? '초대된 프로젝트 구성원' : `${selectedOrganization?.name || '조직'}의 초대된 구성원` }}</dd></div></dl></div>
        <div class="review-group"><div class="review-title"><span class="review-icon">◇</span><div><small>기본 정보</small><b>{{ form.name }}</b></div><button @click="goToStep(1)">수정</button></div><dl><div><dt>Slug</dt><dd class="mono">{{ form.slug }}</dd></div><div><dt>설명</dt><dd>{{ form.description || '설명 없음' }}</dd></div></dl></div>
        <div class="review-group"><div class="review-title"><span class="review-icon">⌘</span><div><small>저장소</small><b>{{ form.repositoryAlias }}</b></div><button @click="goToStep(2)">수정</button></div><dl><div><dt>연결 방식</dt><dd>{{ form.repositorySourceType === 'local_path' ? '서버 경로' : 'Git URL' }}</dd></div><div><dt>위치</dt><dd class="mono break">{{ repositoryLocation }}</dd></div><div><dt>기본 브랜치</dt><dd class="mono">{{ form.defaultBranch }}</dd></div><div><dt>제외 패턴</dt><dd>{{ form.excludePatterns.length }}개</dd></div></dl></div>
        <label class="confirmation"><input v-model="confirmed" type="checkbox"><span>저장소에 대한 읽기 권한이 있으며, 연결 정보와 분석 결과가 프로젝트 구성원에게 공개될 수 있음을 확인했습니다.</span></label>
      </div>

      <footer class="form-actions"><button v-if="currentStep > 0" class="secondary-button" :disabled="submitting" @click="back">← 이전</button><span v-else></span><button v-if="currentStep < steps.length - 1" class="primary-button form-button" @click="next">다음 단계 <span>→</span></button><button v-else class="primary-button form-button" :disabled="submitting" @click="createProject"><span v-if="submitting" class="button-spinner"></span>{{ submitting ? '프로젝트 생성 중…' : '프로젝트 만들기' }}</button></footer>
    </section>
  </div>
</template>
