<template>
  <main class="grid min-h-screen place-items-center px-4 py-8 sm:px-6">
    <section class="premium-surface grid w-full max-w-5xl overflow-hidden rounded-2xl shadow-2xl shadow-brand-950/10 lg:grid-cols-[minmax(0,1fr)_430px]">
      <div class="relative hidden min-h-[560px] overflow-hidden bg-brand-600 p-8 text-white lg:block">
        <div class="absolute inset-0 opacity-20 [background-image:linear-gradient(135deg,rgba(255,255,255,.36)_1px,transparent_1px)] [background-size:28px_28px]" />
        <div class="relative flex h-full flex-col justify-between">
          <div class="grid gap-4">
            <div class="h-1.5 w-16 rounded-full bg-teal-100/80" />
            <div class="max-w-md">
              <slot name="brand" />
            </div>
          </div>
          <div class="grid gap-3 border-t border-white/20 pt-6">
            <div class="h-2 w-3/4 rounded-full bg-white/40" />
            <div class="h-2 w-1/2 rounded-full bg-teal-100/50" />
            <div class="h-2 w-2/3 rounded-full bg-white/25" />
          </div>
        </div>
      </div>
      <div class="p-5 sm:p-8">
        <slot />
      </div>
    </section>
  </main>
</template>
