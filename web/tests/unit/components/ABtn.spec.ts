import { mount } from '@vue/test-utils'

import ABtn from '@/components/ui/ABtn.vue'

describe('ABtn', () => {
  it('renders label and emits click', async () => {
    const wrapper = mount(ABtn, {
      props: {
        label: '提交',
      },
    })

    await wrapper.get('button').trigger('click')

    expect(wrapper.emitted('click')).toBeTruthy()
    expect(wrapper.emitted('click')).toHaveLength(1)
  })

  it('prevents click when loading', async () => {
    const wrapper = mount(ABtn, {
      props: {
        label: '发布',
        loading: true,
        loadingText: '发布中...',
      },
    })

    expect(wrapper.get('button').text()).toContain('发布中...')
    expect(wrapper.get('button').attributes('disabled')).toBeDefined()

    await wrapper.get('button').trigger('click')
    expect(wrapper.emitted('click')).toBeFalsy()
  })
})
